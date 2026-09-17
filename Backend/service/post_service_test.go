package service

import (
	"strings"
	"testing"

	"campusconnect/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// MOCK REPOSITORY (Minggu 5 — Post, Follow, Feed, Search)
// =============================================================================

type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(post *repository.Post) error {
	args := m.Called(post)
	return args.Error(0)
}

func (m *MockPostRepository) FindByID(id string, viewerID string) (*repository.Post, error) {
	args := m.Called(id, viewerID)
	if args.Get(0) != nil {
		return args.Get(0).(*repository.Post), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostRepository) Update(post *repository.Post) error {
	args := m.Called(post)
	return args.Error(0)
}

func (m *MockPostRepository) FindAll(cursor string, limit int, viewerID string) ([]repository.Post, string, error) {
	args := m.Called(cursor, limit, viewerID)
	if args.Get(0) != nil {
		return args.Get(0).([]repository.Post), args.String(1), args.Error(2)
	}
	return nil, args.String(1), args.Error(2)
}

func (m *MockPostRepository) FindByUserIDs(userIDs []string, cursor string, limit int, viewerID string) ([]repository.Post, string, error) {
	args := m.Called(userIDs, cursor, limit, viewerID)
	if args.Get(0) != nil {
		return args.Get(0).([]repository.Post), args.String(1), args.Error(2)
	}
	return nil, args.String(1), args.Error(2)
}

func (m *MockPostRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockFollowRepository struct {
	mock.Mock
}

func (m *MockFollowRepository) FindByFollowerAndFollowing(followerID, followingID string) (*repository.Follow, error) {
	args := m.Called(followerID, followingID)
	if args.Get(0) != nil {
		return args.Get(0).(*repository.Follow), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFollowRepository) Create(follow *repository.Follow) error {
	args := m.Called(follow)
	return args.Error(0)
}

func (m *MockFollowRepository) Delete(followerID, followingID string) error {
	args := m.Called(followerID, followingID)
	return args.Error(0)
}

func (m *MockFollowRepository) FindFollowingIDs(followerID string) ([]string, error) {
	args := m.Called(followerID)
	if args.Get(0) != nil {
		return args.Get(0).([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockSearchRepository struct {
	mock.Mock
}

func (m *MockSearchRepository) SearchUsers(query string, limit int) ([]repository.User, error) {
	args := m.Called(query, limit)
	if args.Get(0) != nil {
		return args.Get(0).([]repository.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSearchRepository) SearchProjects(query string, limit int) ([]repository.Project, error) {
	args := m.Called(query, limit)
	if args.Get(0) != nil {
		return args.Get(0).([]repository.Project), args.Error(1)
	}
	return nil, args.Error(1)
}

// =============================================================================
// POST SERVICE
// =============================================================================

func TestCreatePost_Success(t *testing.T) {
	mockRepo := new(MockPostRepository)
	postService := NewPostService(mockRepo)

	// Repo harus dipanggil dengan konten yang sudah di-trim
	mockRepo.On("Create", mock.AnythingOfType("*repository.Post")).Return(nil)

	post := &repository.Post{UserID: "user-1", Content: "  halo dunia  "}
	err := postService.CreatePost(post)

	assert.NoError(t, err)
	assert.Equal(t, "halo dunia", post.Content)
	mockRepo.AssertExpectations(t)
}

func TestCreatePost_EmptyContent(t *testing.T) {
	mockRepo := new(MockPostRepository)
	postService := NewPostService(mockRepo)

	post := &repository.Post{UserID: "user-1", Content: "   "}
	err := postService.CreatePost(post)

	assert.Error(t, err)
	assert.Equal(t, "konten post tidak boleh kosong", err.Error())
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreatePost_TooLong(t *testing.T) {
	mockRepo := new(MockPostRepository)
	postService := NewPostService(mockRepo)

	post := &repository.Post{UserID: "user-1", Content: strings.Repeat("x", 2001)}
	err := postService.CreatePost(post)

	assert.Error(t, err)
	assert.Equal(t, "konten post maksimal 2000 karakter", err.Error())
	mockRepo.AssertNotCalled(t, "Create")
}

func TestGetPostByID_Passthrough(t *testing.T) {
	mockRepo := new(MockPostRepository)
	postService := NewPostService(mockRepo)

	expected := &repository.Post{ID: "p1", UserID: "user-1", Content: "halo"}
	mockRepo.On("FindByID", "p1", "viewer-1").Return(expected, nil)

	post, err := postService.GetPostByID("p1", "viewer-1")

	assert.NoError(t, err)
	assert.Equal(t, expected, post)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePost_Success(t *testing.T) {
	mockRepo := new(MockPostRepository)
	postService := NewPostService(mockRepo)

	existing := &repository.Post{ID: "p1", UserID: "user-1", Content: "versi lama"}

	// Dipanggil 2x: cek ownership di awal + ambil data terbaru di akhir
	mockRepo.On("FindByID", "p1", "user-1").Return(existing, nil)

	var saved *repository.Post
	mockRepo.On("Update", mock.AnythingOfType("*repository.Post")).
		Run(func(args mock.Arguments) {
			saved = args.Get(0).(*repository.Post)
		}).
		Return(nil)

	updated, err := postService.UpdatePost("p1", "user-1", &repository.Post{Content: "  versi baru  "})

	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "versi baru", updated.Content)
	// Kolom yang dikirim ke repo harus konten baru milik post yang sama
	assert.NotNil(t, saved)
	assert.Equal(t, "p1", saved.ID)
	assert.Equal(t, "versi baru", saved.Content)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePost_Forbidden(t *testing.T) {
	mockRepo := new(MockPostRepository)
	postService := NewPostService(mockRepo)

	existing := &repository.Post{ID: "p1", UserID: "user-2", Content: "milik orang lain"}
	mockRepo.On("FindByID", "p1", "user-1").Return(existing, nil)

	_, err := postService.UpdatePost("p1", "user-1", &repository.Post{Content: "rekayasa"})

	assert.ErrorIs(t, err, repository.ErrForbidden)
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdatePost_NotFound(t *testing.T) {
	mockRepo := new(MockPostRepository)
	postService := NewPostService(mockRepo)

	mockRepo.On("FindByID", "p404", "user-1").Return(nil, repository.ErrNotFound)

	_, err := postService.UpdatePost("p404", "user-1", &repository.Post{Content: "apa pun"})

	assert.ErrorIs(t, err, repository.ErrNotFound)
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdatePost_EmptyContent(t *testing.T) {
	mockRepo := new(MockPostRepository)
	postService := NewPostService(mockRepo)

	existing := &repository.Post{ID: "p1", UserID: "user-1", Content: "lama"}
	mockRepo.On("FindByID", "p1", "user-1").Return(existing, nil)

	_, err := postService.UpdatePost("p1", "user-1", &repository.Post{Content: "   "})

	assert.Error(t, err)
	assert.Equal(t, "konten post tidak boleh kosong", err.Error())
	mockRepo.AssertNotCalled(t, "Update")
}

func TestDeletePost_Success(t *testing.T) {
	mockRepo := new(MockPostRepository)
	postService := NewPostService(mockRepo)

	existing := &repository.Post{ID: "p1", UserID: "user-1", Content: "lama"}
	mockRepo.On("FindByID", "p1", "").Return(existing, nil)
	mockRepo.On("Delete", "p1").Return(nil)

	err := postService.DeletePost("p1", "user-1")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeletePost_Forbidden(t *testing.T) {
	mockRepo := new(MockPostRepository)
	postService := NewPostService(mockRepo)

	existing := &repository.Post{ID: "p1", UserID: "user-2", Content: "lama"}
	mockRepo.On("FindByID", "p1", "").Return(existing, nil)

	err := postService.DeletePost("p1", "user-1")

	assert.ErrorIs(t, err, repository.ErrForbidden)
	mockRepo.AssertNotCalled(t, "Delete")
}

func TestGetAllPosts_LimitClamp(t *testing.T) {
	mockRepo := new(MockPostRepository)
	postService := NewPostService(mockRepo)

	// Limit 999 harus di-clamp ke 50 sebelum menyentuh repository
	mockRepo.On("FindAll", "", 50, "viewer-1").Return([]repository.Post{}, "", nil)
	posts, nextCursor, err := postService.GetAllPosts("", 999, "viewer-1")

	assert.NoError(t, err)
	assert.Empty(t, posts)
	assert.Empty(t, nextCursor)
	mockRepo.AssertExpectations(t)

	// Limit 0 (tidak valid) harus diisi default 10
	mockRepo2 := new(MockPostRepository)
	postService2 := NewPostService(mockRepo2)
	mockRepo2.On("FindAll", "", 10, "viewer-1").Return([]repository.Post{}, "", nil)
	_, _, err = postService2.GetAllPosts("", 0, "viewer-1")
	assert.NoError(t, err)
	mockRepo2.AssertExpectations(t)
}

// =============================================================================
// FOLLOW SERVICE
// =============================================================================

func TestToggleFollow_Follow(t *testing.T) {
	mockFollowRepo := new(MockFollowRepository)
	mockUserRepo := new(MockUserRepository)
	followService := NewFollowService(mockFollowRepo, mockUserRepo)

	target := &repository.User{ID: "user-b", Name: "B"}
	mockUserRepo.On("FindByID", "user-b").Return(target, nil)
	// Belum follow sebelumnya
	mockFollowRepo.On("FindByFollowerAndFollowing", "user-a", "user-b").Return(nil, repository.ErrNotFound)

	var created *repository.Follow
	mockFollowRepo.On("Create", mock.AnythingOfType("*repository.Follow")).
		Run(func(args mock.Arguments) {
			created = args.Get(0).(*repository.Follow)
		}).
		Return(nil)

	following, err := followService.ToggleFollow("user-a", "user-b")

	assert.NoError(t, err)
	assert.True(t, following)
	// Arah relasi harus benar: user-a follower, user-b yang di-follow
	assert.NotNil(t, created)
	assert.Equal(t, "user-a", created.FollowerID)
	assert.Equal(t, "user-b", created.FollowingID)
	mockUserRepo.AssertExpectations(t)
	mockFollowRepo.AssertExpectations(t)
}

func TestToggleFollow_Unfollow(t *testing.T) {
	mockFollowRepo := new(MockFollowRepository)
	mockUserRepo := new(MockUserRepository)
	followService := NewFollowService(mockFollowRepo, mockUserRepo)

	target := &repository.User{ID: "user-b", Name: "B"}
	mockUserRepo.On("FindByID", "user-b").Return(target, nil)
	// Sudah follow sebelumnya
	mockFollowRepo.On("FindByFollowerAndFollowing", "user-a", "user-b").Return(&repository.Follow{FollowerID: "user-a", FollowingID: "user-b"}, nil)
	mockFollowRepo.On("Delete", "user-a", "user-b").Return(nil)

	following, err := followService.ToggleFollow("user-a", "user-b")

	assert.NoError(t, err)
	assert.False(t, following)
	mockFollowRepo.AssertNotCalled(t, "Create")
	mockFollowRepo.AssertExpectations(t)
}

func TestToggleFollow_SelfReject(t *testing.T) {
	mockFollowRepo := new(MockFollowRepository)
	mockUserRepo := new(MockUserRepository)
	followService := NewFollowService(mockFollowRepo, mockUserRepo)

	following, err := followService.ToggleFollow("user-a", "user-a")

	assert.Error(t, err)
	assert.False(t, following)
	assert.Equal(t, "anda tidak bisa follow diri sendiri", err.Error())
	// Tidak boleh menyentuh DB sama sekali
	mockUserRepo.AssertNotCalled(t, "FindByID")
	mockFollowRepo.AssertNotCalled(t, "FindByFollowerAndFollowing")
	mockFollowRepo.AssertNotCalled(t, "Create")
}

func TestToggleFollow_TargetNotFound(t *testing.T) {
	mockFollowRepo := new(MockFollowRepository)
	mockUserRepo := new(MockUserRepository)
	followService := NewFollowService(mockFollowRepo, mockUserRepo)

	mockUserRepo.On("FindByID", "user-ghost").Return(nil, repository.ErrNotFound)

	following, err := followService.ToggleFollow("user-a", "user-ghost")

	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.False(t, following)
	mockFollowRepo.AssertNotCalled(t, "Create")
}

func TestGetUserWithFollowStatus_Following(t *testing.T) {
	mockFollowRepo := new(MockFollowRepository)
	mockUserRepo := new(MockUserRepository)
	followService := NewFollowService(mockFollowRepo, mockUserRepo)

	target := &repository.User{ID: "user-b", Name: "B"}
	mockUserRepo.On("FindByID", "user-b").Return(target, nil)
	mockFollowRepo.On("FindByFollowerAndFollowing", "viewer-1", "user-b").Return(&repository.Follow{}, nil)

	user, following, err := followService.GetUserWithFollowStatus("user-b", "viewer-1")

	assert.NoError(t, err)
	assert.Equal(t, "user-b", user.ID)
	assert.True(t, following)
	mockFollowRepo.AssertExpectations(t)
}

func TestGetUserWithFollowStatus_NotFollowing(t *testing.T) {
	mockFollowRepo := new(MockFollowRepository)
	mockUserRepo := new(MockUserRepository)
	followService := NewFollowService(mockFollowRepo, mockUserRepo)

	target := &repository.User{ID: "user-b", Name: "B"}
	mockUserRepo.On("FindByID", "user-b").Return(target, nil)
	mockFollowRepo.On("FindByFollowerAndFollowing", "viewer-1", "user-b").Return(nil, repository.ErrNotFound)

	user, following, err := followService.GetUserWithFollowStatus("user-b", "viewer-1")

	assert.NoError(t, err)
	assert.Equal(t, "user-b", user.ID)
	assert.False(t, following)
}

func TestGetUserWithFollowStatus_Self(t *testing.T) {
	mockFollowRepo := new(MockFollowRepository)
	mockUserRepo := new(MockUserRepository)
	followService := NewFollowService(mockFollowRepo, mockUserRepo)

	self := &repository.User{ID: "user-a", Name: "A"}
	mockUserRepo.On("FindByID", "user-a").Return(self, nil)

	user, following, err := followService.GetUserWithFollowStatus("user-a", "user-a")

	assert.NoError(t, err)
	assert.Equal(t, "user-a", user.ID)
	assert.False(t, following)
	// Self-view tidak boleh menanyakan relasi follow ke DB
	mockFollowRepo.AssertNotCalled(t, "FindByFollowerAndFollowing")
}

func TestGetUserWithFollowStatus_NotFound(t *testing.T) {
	mockFollowRepo := new(MockFollowRepository)
	mockUserRepo := new(MockUserRepository)
	followService := NewFollowService(mockFollowRepo, mockUserRepo)

	mockUserRepo.On("FindByID", "user-ghost").Return(nil, repository.ErrNotFound)

	user, following, err := followService.GetUserWithFollowStatus("user-ghost", "viewer-1")

	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, user)
	assert.False(t, following)
}

// =============================================================================
// FEED SERVICE
// =============================================================================

func TestGetFeed_IncludesFollowingAndSelf(t *testing.T) {
	mockFollowRepo := new(MockFollowRepository)
	mockPostRepo := new(MockPostRepository)
	feedService := NewFeedService(mockFollowRepo, mockPostRepo)

	// viewer mengikuti user-b; feed = post user-b + post viewer sendiri
	mockFollowRepo.On("FindFollowingIDs", "viewer-1").Return([]string{"user-b"}, nil)

	expectedIDs := []string{"user-b", "viewer-1"}
	mockPostRepo.On("FindByUserIDs", expectedIDs, "", 10, "viewer-1").Return([]repository.Post{}, "", nil)

	posts, nextCursor, err := feedService.GetFeed("viewer-1", "", 10)

	assert.NoError(t, err)
	assert.Empty(t, posts)
	assert.Empty(t, nextCursor)
	mockFollowRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

func TestGetFeed_EmptyFollowingStillSelf(t *testing.T) {
	mockFollowRepo := new(MockFollowRepository)
	mockPostRepo := new(MockPostRepository)
	feedService := NewFeedService(mockFollowRepo, mockPostRepo)

	// Tidak mengikuti siapa pun: feed tetap memuat post milik sendiri
	mockFollowRepo.On("FindFollowingIDs", "viewer-1").Return([]string{}, nil)

	expectedIDs := []string{"viewer-1"}
	mockPostRepo.On("FindByUserIDs", expectedIDs, "", 10, "viewer-1").Return([]repository.Post{}, "", nil)

	_, _, err := feedService.GetFeed("viewer-1", "", 10)

	assert.NoError(t, err)
	mockFollowRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

func TestGetFeed_LimitClamp(t *testing.T) {
	mockFollowRepo := new(MockFollowRepository)
	mockPostRepo := new(MockPostRepository)
	feedService := NewFeedService(mockFollowRepo, mockPostRepo)

	mockFollowRepo.On("FindFollowingIDs", "viewer-1").Return([]string{}, nil)
	// Limit 100 harus di-clamp ke 50
	mockPostRepo.On("FindByUserIDs", []string{"viewer-1"}, "", 50, "viewer-1").Return([]repository.Post{}, "", nil)

	_, _, err := feedService.GetFeed("viewer-1", "", 100)

	assert.NoError(t, err)
	mockPostRepo.AssertExpectations(t)
}

// =============================================================================
// SEARCH SERVICE
// =============================================================================

func TestSearch_Success(t *testing.T) {
	mockRepo := new(MockSearchRepository)
	searchService := NewSearchService(mockRepo)

	mockRepo.On("SearchUsers", "ali", 10).Return([]repository.User{{ID: "u1", Name: "Ali Saroji"}}, nil)
	mockRepo.On("SearchProjects", "ali", 10).Return([]repository.Project{{ID: "pr1", Title: "Kalkulator Ali"}}, nil)

	result, err := searchService.Search("  ali  ", 10)

	assert.NoError(t, err)
	assert.Len(t, result.Users, 1)
	assert.Equal(t, "Ali Saroji", result.Users[0].Name)
	assert.Len(t, result.Projects, 1)
	assert.Equal(t, "Kalkulator Ali", result.Projects[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestSearch_EmptyQuery(t *testing.T) {
	mockRepo := new(MockSearchRepository)
	searchService := NewSearchService(mockRepo)

	result, err := searchService.Search("   ", 10)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "query pencarian tidak boleh kosong", err.Error())
	mockRepo.AssertNotCalled(t, "SearchUsers")
	mockRepo.AssertNotCalled(t, "SearchProjects")
}

func TestSearch_NilNormalizedToEmptyArray(t *testing.T) {
	mockRepo := new(MockSearchRepository)
	searchService := NewSearchService(mockRepo)

	// Repo mengembalikan nil (tidak ada hasil): service harus menormalkan
	// menjadi array kosong agar frontend tidak menerima null.
	mockRepo.On("SearchUsers", "zzz", 10).Return(nil, nil)
	mockRepo.On("SearchProjects", "zzz", 10).Return(nil, nil)

	result, err := searchService.Search("zzz", 10)

	assert.NoError(t, err)
	assert.NotNil(t, result.Users)
	assert.NotNil(t, result.Projects)
	assert.Len(t, result.Users, 0)
	assert.Len(t, result.Projects, 0)
}

func TestSearch_LimitClamp(t *testing.T) {
	mockRepo := new(MockSearchRepository)
	searchService := NewSearchService(mockRepo)

	// Limit 100 harus di-clamp ke 25
	mockRepo.On("SearchUsers", "go", 25).Return(nil, nil)
	mockRepo.On("SearchProjects", "go", 25).Return(nil, nil)

	_, err := searchService.Search("go", 100)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSearch_RepoErrorPropagates(t *testing.T) {
	mockRepo := new(MockSearchRepository)
	searchService := NewSearchService(mockRepo)

	mockRepo.On("SearchUsers", "ali", 10).Return(nil, assert.AnError)

	result, err := searchService.Search("ali", 10)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertNotCalled(t, "SearchProjects")
}
