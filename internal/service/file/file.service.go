package file

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/isd-sgcu/rpkm66-file/cfgldr"
	fileConst "github.com/isd-sgcu/rpkm66-file/constant/file"
	dto "github.com/isd-sgcu/rpkm66-file/internal/dto/file"
	entity "github.com/isd-sgcu/rpkm66-file/internal/entity/file"
	proto "github.com/isd-sgcu/rpkm66-file/internal/proto/rpkm66/file/file/v1"
	"github.com/isd-sgcu/rpkm66-file/internal/utils"
	"github.com/isd-sgcu/rpkm66-file/pkg/client/gcs"
	"github.com/isd-sgcu/rpkm66-file/pkg/repository/cache"
	"github.com/isd-sgcu/rpkm66-file/pkg/repository/file"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serviceImpl struct {
	proto.UnimplementedFileServiceServer
	conf       cfgldr.GCS
	ttl        int
	client     gcs.Client
	repository file.Repository
	cacheRepo  cache.CacheRepository
}

func NewService(conf cfgldr.GCS, ttl int, client gcs.Client, repository file.Repository, cacheRepo cache.CacheRepository) *serviceImpl {
	return &serviceImpl{
		conf:       conf,
		ttl:        ttl,
		client:     client,
		repository: repository,
		cacheRepo:  cacheRepo,
	}
}

func (s *serviceImpl) Upload(_ context.Context, req *proto.UploadRequest) (*proto.UploadResponse, error) {
	if req.Data == nil {
		return nil, status.Error(codes.InvalidArgument, "File cannot be empty")
	}

	filename, err := utils.GetObjectName(req.Filename, s.conf.Secret, fileConst.Type(req.Type))
	if err != nil {
		log.Error().Err(err).
			Str("service", "file").
			Str("module", "upload image").
			Str("file_name", filename).
			Msg("Invalid file type")
		return nil, status.Error(codes.InvalidArgument, "Invalid file type")
	}

	err = s.client.Upload(req.Data, filename)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Msg("Cannot connect to google cloud storage")
		return nil, status.Error(codes.Unavailable, "Cannot connect to google cloud storage")
	}

	f := &entity.File{
		Filename: filename,
		OwnerID:  req.UserId,
		Tag:      int(req.Tag),
	}

	err = s.repository.CreateOrUpdate(f)

	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Str("filename", filename).
			Str("user_id", req.UserId).
			Msg("Error while saving file data")
		return nil, status.Error(codes.Unavailable, "Internal service error")
	}

	url, err := s.client.GetSignedUrl(filename)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Str("filename", filename).
			Str("user_id", req.UserId).
			Msg("Error while trying to get signed url")
		return nil, status.Error(codes.Unavailable, "Internal service error")
	}

	cacheFile := dto.CacheFile{
		Url:      url,
		Filename: filename,
	}

	err = s.cacheRepo.SaveCache(req.UserId, &cacheFile, s.ttl)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Str("filename", filename).
			Str("user_id", req.UserId).
			Interface("cache", cacheFile).
			Msg("Error while connecting to redis server")
		return nil, status.Error(codes.Unavailable, "Error while connecting to redis server")
	}

	return &proto.UploadResponse{Url: url}, nil
}

func (s *serviceImpl) GetSignedUrl(_ context.Context, req *proto.GetSignedUrlRequest) (*proto.GetSignedUrlResponse, error) {
	cachedFile := &dto.CacheFile{}
	err := s.cacheRepo.GetCache(req.UserId, cachedFile)
	if err == nil {
		return &proto.GetSignedUrlResponse{Url: cachedFile.Url}, nil
	}

	if err != redis.Nil {
		log.Error().
			Err(err).
			Str("module", "get signed url").
			Str("user_id", req.UserId).
			Msg("Error while connecting to redis server")
		return nil, status.Error(codes.Unavailable, "Error while connecting to redis server")
	}

	f := entity.File{}
	err = s.repository.FindByOwnerID(req.UserId, &f)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "get signed url").
			Str("user_id", req.UserId).
			Msg("Error while trying to query data")
		return nil, status.Error(codes.NotFound, "Not found file")
	}

	url, err := s.client.GetSignedUrl(f.Filename)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Str("filename", cachedFile.Filename).
			Str("user_id", req.UserId).
			Msg("Cannot connect to google cloud storage")
		return nil, status.Error(codes.Unavailable, "Cannot connect to google cloud storage")
	}

	cachedFile = &dto.CacheFile{
		Url:      url,
		Filename: f.Filename,
	}

	err = s.cacheRepo.SaveCache(req.UserId, cachedFile, s.ttl)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload file").
			Str("filename", cachedFile.Filename).
			Str("user_id", req.UserId).
			Interface("cache", cachedFile).
			Msg("Error while connecting to redis server")
		return nil, status.Error(codes.Unavailable, "Error while connecting to redis server")
	}

	return &proto.GetSignedUrlResponse{Url: url}, nil
}
