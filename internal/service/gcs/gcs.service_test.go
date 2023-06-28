package gcs

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/bxcodec/faker/v4"
	"github.com/go-redis/redis/v8"
	"github.com/isd-sgcu/rpkm66-file/cfgldr"
	dto "github.com/isd-sgcu/rpkm66-file/internal/dto/file"
	"github.com/isd-sgcu/rpkm66-file/internal/model/file"
	proto "github.com/isd-sgcu/rpkm66-file/internal/proto/rpkm66/file/file/v1"
	cMock "github.com/isd-sgcu/rpkm66-file/mocks/cache"
	fMock "github.com/isd-sgcu/rpkm66-file/mocks/file"
	mock "github.com/isd-sgcu/rpkm66-file/mocks/gcs"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GCSServiceTest struct {
	suite.Suite
	conf      cfgldr.GCS
	filename  string
	file      []byte
	url       string
	err       error
	f         *file.File
	ttl       int
	cacheFile *dto.CacheFile
}

func TestGCSService(t *testing.T) {
	suite.Run(t, new(GCSServiceTest))
}

func (t *GCSServiceTest) SetupTest() {
	t.filename = fmt.Sprintf("file-%s", faker.Word())

	t.url = faker.URL()

	t.conf = cfgldr.GCS{
		BucketName:          faker.Word(),
		Secret:              faker.Word(),
		ServiceAccountKey:   []byte(faker.Word()),
		ServiceAccountEmail: faker.Word(),
	}

	t.f = &file.File{
		Filename: t.filename,
		OwnerID:  faker.UUIDDigit(),
		Tag:      1,
	}

	t.file = []byte("Hello")

	t.err = errors.New("Something wrong :(")

	t.ttl = 15 * 60

	t.cacheFile = &dto.CacheFile{
		Url:      t.url,
		Filename: t.filename,
	}
}

func (t *GCSServiceTest) TestUploadSuccess() {
	want := &proto.UploadResponse{Url: t.url}

	c := mock.ClientMock{}
	c.On("Upload", t.file).Return(nil)
	c.On("GetSignedUrl").Return(t.url, nil)

	repo := fMock.RepositoryMock{}
	repo.On("CreateOrUpdate", t.f.OwnerID).Return(t.f, nil)

	cacheRepo := cMock.RepositoryMock{V: map[string]interface{}{}}
	cacheRepo.On("SaveCache", t.f.OwnerID, t.cacheFile.Url, t.ttl).Return(nil)

	srv := NewService(t.conf, t.ttl, &c, &repo, &cacheRepo)

	actual, err := srv.Upload(context.Background(), &proto.UploadRequest{
		Filename: t.filename,
		Data:     t.file,
		UserId:   t.f.OwnerID,
		Tag:      1,
	})

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, actual)
}

func (t *GCSServiceTest) TestUploadFailed() {
	c := mock.ClientMock{}
	c.On("Upload", t.file).Return(errors.New("Cannot upload file"))
	c.On("GetSignedUrl").Return(t.url, nil)

	repo := fMock.RepositoryMock{}
	repo.On("CreateOrUpdate", t.f.OwnerID).Return(t.f, nil)

	cacheRepo := cMock.RepositoryMock{V: map[string]interface{}{}}
	cacheRepo.On("SaveCache", t.f.OwnerID, t.cacheFile.Url, t.ttl).Return(nil)

	srv := NewService(t.conf, t.ttl, &c, &repo, &cacheRepo)

	actual, err := srv.Upload(context.Background(), &proto.UploadRequest{
		Filename: t.filename,
		Data:     t.file,
		UserId:   t.f.OwnerID,
		Tag:      1,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}

func (t *GCSServiceTest) TestGetSignedUrlCachedSuccess() {
	t.f.Filename = fmt.Sprintf("%s-%s", t.filename, faker.Word())
	str := strings.Split(t.f.Filename, "file-")
	t.filename = str[1]

	want := &proto.GetSignedUrlResponse{Url: t.url}

	c := mock.ClientMock{}

	repo := fMock.RepositoryMock{}

	cacheRepo := cMock.RepositoryMock{V: map[string]interface{}{}}
	cacheRepo.On("GetCache", t.f.OwnerID, &dto.CacheFile{}).Return(t.cacheFile, nil)

	srv := NewService(t.conf, t.ttl, &c, &repo, &cacheRepo)

	actual, err := srv.GetSignedUrl(context.Background(), &proto.GetSignedUrlRequest{
		UserId: t.f.OwnerID,
	})

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, actual)
}
func (t *GCSServiceTest) TestGetSignedUrlCachedErr() {
	c := mock.ClientMock{}

	repo := fMock.RepositoryMock{}

	cacheRepo := cMock.RepositoryMock{V: map[string]interface{}{}}
	cacheRepo.On("GetCache", t.f.OwnerID, &dto.CacheFile{}).Return(nil, errors.New("Cannot connect to redis server"))

	srv := NewService(t.conf, t.ttl, &c, &repo, &cacheRepo)

	actual, err := srv.GetSignedUrl(context.Background(), &proto.GetSignedUrlRequest{
		UserId: t.f.OwnerID,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}

func (t *GCSServiceTest) TestGetSignedUrlSuccessSaveCacheSuccess() {
	want := &proto.GetSignedUrlResponse{Url: t.url}

	c := mock.ClientMock{}
	c.On("GetSignedUrl").Return(t.url, nil)

	repo := fMock.RepositoryMock{}
	repo.On("FindByOwnerID", t.f.OwnerID, &file.File{}).Return(t.f, nil)

	cacheRepo := cMock.RepositoryMock{V: map[string]interface{}{}}
	cacheRepo.On("GetCache", t.f.OwnerID, &dto.CacheFile{}).Return(nil, redis.Nil)
	cacheRepo.On("SaveCache", t.f.OwnerID, t.cacheFile.Url, t.ttl).Return(nil)

	srv := NewService(t.conf, t.ttl, &c, &repo, &cacheRepo)

	actual, err := srv.GetSignedUrl(context.Background(), &proto.GetSignedUrlRequest{
		UserId: t.f.OwnerID,
	})

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, actual)
}

func (t *GCSServiceTest) TestGetSignedUrlSuccessSaveCacheFailed() {
	c := mock.ClientMock{}
	c.On("GetSignedUrl").Return(t.url, nil)

	repo := fMock.RepositoryMock{}
	repo.On("FindByOwnerID", t.f.OwnerID, &file.File{}).Return(t.f, nil)

	cacheRepo := cMock.RepositoryMock{V: map[string]interface{}{}}
	cacheRepo.On("GetCache", t.f.OwnerID, &dto.CacheFile{}).Return(nil, redis.Nil)
	cacheRepo.On("SaveCache", t.f.OwnerID, t.cacheFile.Url, t.ttl).Return(errors.New("Cannot connect to redis server"))

	srv := NewService(t.conf, t.ttl, &c, &repo, &cacheRepo)

	actual, err := srv.GetSignedUrl(context.Background(), &proto.GetSignedUrlRequest{
		UserId: t.f.OwnerID,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}

func (t *GCSServiceTest) TestGetSignedUrlFailed() {
	c := mock.ClientMock{}
	c.On("GetSignedUrl").Return("", t.err)

	repo := fMock.RepositoryMock{}
	repo.On("FindByOwnerID", t.f.OwnerID, &file.File{}).Return(t.f, nil)

	cacheRepo := cMock.RepositoryMock{V: map[string]interface{}{}}
	cacheRepo.On("GetCache", t.f.OwnerID, &dto.CacheFile{}).Return(nil, redis.Nil)
	srv := NewService(t.conf, t.ttl, &c, &repo, &cacheRepo)

	actual, err := srv.GetSignedUrl(context.Background(), &proto.GetSignedUrlRequest{
		UserId: t.f.OwnerID,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}

func (t *GCSServiceTest) TestGetSignedUrlNotFound() {
	c := mock.ClientMock{}
	c.On("GetSignedUrl").Return("", t.err)

	repo := fMock.RepositoryMock{}
	repo.On("FindByOwnerID", t.f.OwnerID, &file.File{}).Return(nil, errors.New("Not found file"))

	cacheRepo := cMock.RepositoryMock{V: map[string]interface{}{}}
	cacheRepo.On("GetCache", t.f.OwnerID, &dto.CacheFile{}).Return(nil, redis.Nil)

	srv := NewService(t.conf, t.ttl, &c, &repo, &cacheRepo)

	actual, err := srv.GetSignedUrl(context.Background(), &proto.GetSignedUrlRequest{
		UserId: t.f.OwnerID,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.NotFound, st.Code())
}
