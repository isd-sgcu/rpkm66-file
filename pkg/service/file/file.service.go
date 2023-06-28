package file

import (
	"context"

	"github.com/isd-sgcu/rpkm66-file/cfgldr"
	proto "github.com/isd-sgcu/rpkm66-file/internal/proto/rpkm66/file/file/v1"
	fileSvc "github.com/isd-sgcu/rpkm66-file/internal/service/file"
	"github.com/isd-sgcu/rpkm66-file/pkg/client/gcs"
	"github.com/isd-sgcu/rpkm66-file/pkg/repository/cache"
	"github.com/isd-sgcu/rpkm66-file/pkg/repository/file"
)

type FileService interface {
	proto.FileServiceServer
	Upload(_ context.Context, req *proto.UploadRequest) (*proto.UploadResponse, error)
	GetSignedUrl(_ context.Context, req *proto.GetSignedUrlRequest) (*proto.GetSignedUrlResponse, error)
}

func NewFileService(conf cfgldr.GCS, ttl int, client gcs.Client, repository file.Repository, cacheRepo cache.CacheRepository) FileService {
	return fileSvc.NewService(conf, ttl, client, repository, cacheRepo)
}
