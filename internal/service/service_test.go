package service

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/istvzsig/artifix/internal/artifact"
	"github.com/istvzsig/artifix/internal/repository"
)

type fakeArtifactRepository struct {
	created   bool
	createErr error
	updated   bool
	updateErr error
	deleted   bool
	deleteErr error
	calls     *[]string
}

var _ repository.ArtifactRepository = (*fakeArtifactRepository)(nil)

func (f *fakeArtifactRepository) Create(
	ctx context.Context,
	a artifact.Artifact,
) error {
	f.created = true
	recordCall(f.calls, "Create")

	if f.createErr != nil {
		return f.createErr
	}

	return nil
}

func (f *fakeArtifactRepository) Get(
	ctx context.Context,
	coords artifact.Coordinates,
) (artifact.Artifact, error) {
	return artifact.Artifact{}, nil
}

func (f *fakeArtifactRepository) Update(
	ctx context.Context,
	a artifact.Artifact,
) error {
	f.updated = true
	recordCall(f.calls, "Update")

	if f.updateErr != nil {
		return f.updateErr
	}

	return nil
}

func (f *fakeArtifactRepository) List(
	ctx context.Context,
	namespace string,
	name string,
) ([]artifact.Artifact, error) {
	return []artifact.Artifact{}, nil
}

func (f *fakeArtifactRepository) Delete(
	ctx context.Context,
	coords artifact.Coordinates,
) error {
	f.deleted = true
	recordCall(f.calls, "Delete")

	if f.deleteErr != nil {
		return f.deleteErr
	}

	return nil
}

type fakeBinaryStorage struct {
	published       bool
	publishErr      error
	discardedStage  bool
	discardStageErr error
	calls           *[]string
}

func (f *fakeBinaryStorage) Stage(
	ctx context.Context,
	content io.Reader,
) (artifact.Digest, int64, error) {
	recordCall(f.calls, "Stage")
	return artifact.Digest{}, 14, nil
}

func (f *fakeBinaryStorage) Publish(
	ctx context.Context,
	digest artifact.Digest,
) error {
	f.published = true
	recordCall(f.calls, "Publish")

	if f.publishErr != nil {
		return f.publishErr
	}

	return nil
}

func (f *fakeBinaryStorage) DiscardStage(
	ctx context.Context,
	digest artifact.Digest,
) error {
	f.discardedStage = true
	recordCall(f.calls, "DiscardStage")

	if f.discardStageErr != nil {
		return f.discardStageErr
	}

	return nil
}

func recordCall(calls *[]string, name string) {
	if calls != nil {
		*calls = append(*calls, name)
	}
}

func TestNewService_Success(t *testing.T) {
	repository := &fakeArtifactRepository{}
	storage := &fakeBinaryStorage{}

	_, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewService_NilRepository(t *testing.T) {
	storage := &fakeBinaryStorage{}

	_, err := NewService(nil, storage)
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected ErrRepositoryRequired, got %v", err)
	}
}

func TestNewService_NilStorage(t *testing.T) {
	repository := &fakeArtifactRepository{}

	_, err := NewService(repository, nil)
	if !errors.Is(err, ErrBinaryStorageRequired) {
		t.Fatalf("expected ErrBinaryStorageRequired, got %v", err)
	}
}

func TestService_Upload_Success(t *testing.T) {
	repository := &fakeArtifactRepository{}
	storage := &fakeBinaryStorage{}

	service, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	got, err := service.Upload(
		t.Context(),
		coords,
		strings.NewReader("test artifact"),
		"application/zip",
	)
	if err != nil {
		t.Fatal(err)
	}

	if got.Size() != 14 {
		t.Errorf("expected size 14, got %d", got.Size())
	}

	if got.Digest() != (artifact.Digest{}) {
		t.Errorf("expected zero digest, got %v", got.Digest())
	}

	if got.Status() != artifact.StatusAvailable {
		t.Errorf("expected status AVAILABLE, got %s", got.Status())
	}

	if !repository.created {
		t.Error("expected repository.Create to be called")
	}

	if !repository.updated {
		t.Error("expected repository.Update to be called")
	}

	if !storage.published {
		t.Error("expected storage.Publish to be called")
	}
}

func TestService_Upload_CreateFailure(t *testing.T) {
	repository := &fakeArtifactRepository{
		createErr: ErrDatabaseUnavailable,
	}
	storage := &fakeBinaryStorage{}

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Upload(
		t.Context(),
		coords,
		strings.NewReader("test artifact"),
		"application/zip",
	)

	if !errors.Is(err, ErrDatabaseUnavailable) {
		t.Fatalf("expected ErrDatabaseUnavailable, got %v", err)
	}

	if !repository.created {
		t.Error("expected repository.Create to be called")
	}

	if storage.published {
		t.Error("expected storage.Publish not to be called")
	}

	if !storage.discardedStage {
		t.Error("expected storage.DiscardStage to be called")
	}

	if repository.updated {
		t.Error("expected repository.Update not to be called")
	}
}

func TestService_Upload_PublishFailure(t *testing.T) {
	repository := &fakeArtifactRepository{}
	storage := &fakeBinaryStorage{
		publishErr: ErrBinaryStoragePublishFailed,
	}

	service, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Upload(
		t.Context(),
		coords,
		strings.NewReader("test artifact"),
		"application/zip",
	)

	if !errors.Is(err, ErrBinaryStoragePublishFailed) {
		t.Fatalf("expected ErrBinaryStoragePublishFailed, got %v", err)
	}

	if !repository.created {
		t.Error("expected repository.Create to be called")
	}

	if !storage.published {
		t.Error("expected storage.Publish to be called")
	}

	if !storage.discardedStage {
		t.Error("expected storage.DiscardStage to be called")
	}

	if repository.updated {
		t.Error("expected repository.Update not to be called")
	}
}

func TestService_Upload_UpdateFailure(t *testing.T) {
	repository := &fakeArtifactRepository{
		updateErr: ErrRepositoryUpdateFailed,
	}
	storage := &fakeBinaryStorage{}

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Upload(
		t.Context(),
		coords,
		strings.NewReader("test artifact"),
		"application/zip",
	)

	if !errors.Is(err, ErrRepositoryUpdateFailed) {
		t.Fatalf("expected ErrRepositoryUpdateFailed, got %v", err)
	}

	if !storage.published {
		t.Error("expected storage.Publish to be called")
	}

	if !repository.created {
		t.Error("expected repository.Create to be called")
	}

	if !repository.updated {
		t.Error("expected repository.Update to be called")
	}

	if storage.discardedStage {
		t.Error("expected published binary not to be discarded")
	}
}

func TestService_Upload_CreateFailureAndCleanupFailure(t *testing.T) {
	repository := &fakeArtifactRepository{
		createErr: ErrDatabaseUnavailable,
	}
	storage := &fakeBinaryStorage{
		discardStageErr: ErrBinaryStorageDiscardStageFailed,
	}

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Upload(
		t.Context(),
		coords,
		strings.NewReader("test artifact"),
		"application/zip",
	)

	if !errors.Is(err, ErrDatabaseUnavailable) {
		t.Fatalf("expected ErrDatabaseUnavailable, got %v", err)
	}

	if !repository.created {
		t.Error("expected repository.Create to be called")
	}

	if !storage.discardedStage {
		t.Error("expected storage.DiscardStage to be called")
	}

	if storage.published {
		t.Error("expected storage.Publish to be called")
	}
}

func TestService_Upload_InvalidMetadata(t *testing.T) {
	repository := &fakeArtifactRepository{}
	storage := &fakeBinaryStorage{}

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Upload(
		t.Context(),
		coords,
		strings.NewReader("test artifact"),
		"not-a-valid-media-type",
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, artifact.ErrInvalidMediaType) {
		t.Fatalf("expected artifact.ErrInvalidMediaType, got %v", err)
	}

	if !storage.discardedStage {
		t.Error("expected storage.DiscardStage to be called")
	}

	if repository.created {
		t.Error("expected repository.Create not to be called")
	}

	if storage.published {
		t.Error("expected storage.Publish not to be called")
	}

	if repository.updated {
		t.Error("expected repository.Update not to be called")
	}
}

func TestService_Operations_SuccessOrdering(t *testing.T) {
	calls := []string{}

	repository := &fakeArtifactRepository{
		calls: &calls,
	}
	storage := &fakeBinaryStorage{
		calls: &calls,
	}

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Upload(
		t.Context(),
		coords,
		strings.NewReader("test artifact"),
		"application/zip",
	)
	if err != nil {
		t.Fatal(err)
	}

	wantCalls := []string{"Stage", "Create", "Publish", "Update"}

	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls %v ordering does not match with %v", calls, wantCalls)
	}
}

func TestService_Operations_CreateFailureOrdering(t *testing.T) {
	calls := []string{}

	repository := &fakeArtifactRepository{
		createErr: ErrDatabaseUnavailable,
		calls:     &calls,
	}
	storage := &fakeBinaryStorage{
		calls: &calls,
	}

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Upload(
		t.Context(),
		coords,
		strings.NewReader("test artifact"),
		"application/zip",
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrDatabaseUnavailable) {
		t.Fatalf("expected ErrDatabaseUnavailable, got %v", err)
	}

	wantCalls := []string{"Stage", "Create", "DiscardStage"}

	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls %v ordering does not match with %v", calls, wantCalls)
	}
}

func TestService_Operations_PublishFailureOrdering(t *testing.T) {
	calls := []string{}

	repository := &fakeArtifactRepository{
		calls: &calls,
	}
	storage := &fakeBinaryStorage{
		publishErr: ErrBinaryStoragePublishFailed,
		calls:      &calls,
	}

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Upload(
		t.Context(),
		coords,
		strings.NewReader("test artifact"),
		"application/zip",
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrBinaryStoragePublishFailed) {
		t.Fatalf("expected ErrBinaryStoragePublishFailed, got %v", err)
	}

	wantCalls := []string{"Stage", "Create", "Publish", "DiscardStage"}

	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls %v ordering does not match with %v", calls, wantCalls)
	}
}

func TestService_Operations_UpdateFailureOrdering(t *testing.T) {
	calls := []string{}

	repository := &fakeArtifactRepository{
		updateErr: ErrRepositoryUpdateFailed,
		calls:     &calls,
	}
	storage := &fakeBinaryStorage{
		calls: &calls,
	}

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Upload(
		t.Context(),
		coords,
		strings.NewReader("test artifact"),
		"application/zip",
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrRepositoryUpdateFailed) {
		t.Fatalf("expected ErrRepositoryUpdateFailed, got %v", err)
	}

	wantCalls := []string{"Stage", "Create", "Publish", "Update"}

	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls %v ordering does not match with %v", calls, wantCalls)
	}
}
func TestService_Delete_Success(t *testing.T) {
	calls := []string{}

	repository := &fakeArtifactRepository{
		calls: &calls,
	}
	storage := &fakeBinaryStorage{}

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}

	err = service.Delete(t.Context(), coords)
	if err != nil {
		t.Fatalf("expected no delete error, got %v", err)
	}

	if !repository.deleted {
		t.Error("expected repository.Delete to be called")
	}

	if storage.published {
		t.Error("expected storage.Publish not to be called")
	}

	if !reflect.DeepEqual(calls, []string{"Delete"}) {
		t.Fatalf("expected calls [Delete], got %v", calls)
	}
}

func TestService_Delete_RepositoryFailure(t *testing.T) {
	repository := &fakeArtifactRepository{
		deleteErr: ErrDatabaseUnavailable,
	}
	storage := &fakeBinaryStorage{}

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}

	err = service.Delete(t.Context(), coords)
	if !errors.Is(err, ErrDatabaseUnavailable) {
		t.Fatalf("expected error %v, got %v", ErrDatabaseUnavailable, err)
	}
}
