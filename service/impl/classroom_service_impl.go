package impl

import (
	"context"
	"crypto/rand"
	"errors"
	"evasbr/mclamg/common"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/exception"
	"evasbr/mclamg/model"
	"evasbr/mclamg/repository"
	"evasbr/mclamg/service"
	"fmt"
	"math/big"
	"mime/multipart"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type classroomServiceImpl struct {
	ClassroomRepository repository.ClassroomRepository
	Storage             common.FileStorage
	log                 *logrus.Entry
}

func NewClassroomServiceImpl(classroomRepository *repository.ClassroomRepository, storage common.FileStorage) service.ClassroomService {
	return &classroomServiceImpl{
		ClassroomRepository: *classroomRepository,
		Storage:             storage,
		log:                 common.Log.WithField("scope", "ClassroomService"),
	}
}

func generateInviteCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}

func (s *classroomServiceImpl) Create(ctx context.Context, request model.CreateClassroomRequest, coverHeader *multipart.FileHeader, wallpaperHeader *multipart.FileHeader, creatorID string) (model.ClassroomResponse, error) {
	common.Validate(request)

	creatorUUID, err := uuid.Parse(creatorID)
	if err != nil {
		return model.ClassroomResponse{}, exception.ValidationError{
			Message: "invalid creator ID format",
		}
	}

	// Generate unique invite code (check uniqueness)
	var inviteCode string
	for i := 0; i < 5; i++ {
		code, err := generateInviteCode()
		if err != nil {
			return model.ClassroomResponse{}, err
		}
		// check if code exists
		_, err = s.ClassroomRepository.FindByCode(ctx, code)
		if err != nil {
			// Code not found, we can use it
			inviteCode = code
			break
		}
	}
	if inviteCode == "" {
		return model.ClassroomResponse{}, errors.New("failed to generate unique invite code")
	}

	isInviteEnable := true
	if request.IsExternalInviteEnable != nil {
		isInviteEnable = *request.IsExternalInviteEnable
	}

	classroomID := uuid.New()

	var coverImg *string
	if coverHeader != nil {
		file, err := coverHeader.Open()
		if err != nil {
			return model.ClassroomResponse{}, exception.ValidationError{
				Message: fmt.Sprintf("unable to open cover file stream: %v", err),
			}
		}
		defer file.Close()
		filename := fmt.Sprintf("%s_cover_%d", classroomID.String(), time.Now().Unix())
		url, err := s.Storage.UploadFile(ctx, file, filename, common.FolderClassrooms)
		if err != nil {
			return model.ClassroomResponse{}, err
		}
		coverImg = &url
	}

	var wallpaperImg *string
	if wallpaperHeader != nil {
		file, err := wallpaperHeader.Open()
		if err != nil {
			return model.ClassroomResponse{}, exception.ValidationError{
				Message: fmt.Sprintf("unable to open wallpaper file stream: %v", err),
			}
		}
		defer file.Close()
		filename := fmt.Sprintf("%s_wallpaper_%d", classroomID.String(), time.Now().Unix())
		url, err := s.Storage.UploadFile(ctx, file, filename, common.FolderClassrooms)
		if err != nil {
			// Cleanup cover image if uploaded
			if coverImg != nil {
				_ = s.Storage.DeleteFile(ctx, *coverImg)
			}
			return model.ClassroomResponse{}, err
		}
		wallpaperImg = &url
	}

	classroomEntity := entity.Classroom{
		ID:                     classroomID,
		Name:                   request.Name,
		Description:            request.Description,
		Codes:                  inviteCode,
		CoverImg:               coverImg,
		WallpaperImg:           wallpaperImg,
		IsExternalInviteEnable: isInviteEnable,
		Status:                 "ACTIVE",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	created, err := s.ClassroomRepository.Create(ctx, classroomEntity, creatorUUID)
	if err != nil {
		// Cleanup uploaded images on DB failure
		if coverImg != nil {
			_ = s.Storage.DeleteFile(ctx, *coverImg)
		}
		if wallpaperImg != nil {
			_ = s.Storage.DeleteFile(ctx, *wallpaperImg)
		}
		return model.ClassroomResponse{}, err
	}

	return s.toClassroomResponse(created, true), nil
}

func (s *classroomServiceImpl) Update(ctx context.Context, request model.UpdateClassroomRequest, coverHeader *multipart.FileHeader, wallpaperHeader *multipart.FileHeader, id string) (model.ClassroomResponse, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return model.ClassroomResponse{}, exception.ValidationError{
			Message: "invalid classroom ID format",
		}
	}

	common.Validate(request)

	existing, err := s.ClassroomRepository.FindByID(ctx, parsedUUID)
	if err != nil {
		return model.ClassroomResponse{}, exception.NotFoundError{
			Message: err.Error(),
		}
	}

	isInviteEnable := existing.IsExternalInviteEnable
	if request.IsExternalInviteEnable != nil {
		isInviteEnable = *request.IsExternalInviteEnable
	}

	if request.Name != nil {
		existing.Name = *request.Name
	}
	if request.Description != nil {
		existing.Description = request.Description
	}
	existing.IsExternalInviteEnable = isInviteEnable
	if request.Status != nil {
		existing.Status = *request.Status
	}
	existing.UpdatedAt = time.Now()

	if coverHeader != nil {
		file, err := coverHeader.Open()
		if err != nil {
			return model.ClassroomResponse{}, exception.ValidationError{
				Message: fmt.Sprintf("unable to open cover file stream: %v", err),
			}
		}
		defer file.Close()
		filename := fmt.Sprintf("%s_cover_%d", existing.ID.String(), time.Now().Unix())
		url, err := s.Storage.UploadFile(ctx, file, filename, common.FolderClassrooms)
		if err != nil {
			return model.ClassroomResponse{}, err
		}
		newCoverImg := &url

		// Delete old cover from storage
		if existing.CoverImg != nil && *existing.CoverImg != "" {
			_ = s.Storage.DeleteFile(ctx, *existing.CoverImg)
		}
		existing.CoverImg = newCoverImg
	}

	if wallpaperHeader != nil {
		file, err := wallpaperHeader.Open()
		if err != nil {
			return model.ClassroomResponse{}, exception.ValidationError{
				Message: fmt.Sprintf("unable to open wallpaper file stream: %v", err),
			}
		}
		defer file.Close()
		filename := fmt.Sprintf("%s_wallpaper_%d", existing.ID.String(), time.Now().Unix())
		url, err := s.Storage.UploadFile(ctx, file, filename, common.FolderClassrooms)
		if err != nil {
			return model.ClassroomResponse{}, err
		}
		newWallpaperImg := &url

		// Delete old wallpaper from storage
		if existing.WallpaperImg != nil && *existing.WallpaperImg != "" {
			_ = s.Storage.DeleteFile(ctx, *existing.WallpaperImg)
		}
		existing.WallpaperImg = newWallpaperImg
	}

	updated, err := s.ClassroomRepository.Update(ctx, existing)
	if err != nil {
		return model.ClassroomResponse{}, err
	}

	return s.toClassroomResponse(updated, true), nil
}

func (s *classroomServiceImpl) Delete(ctx context.Context, id string) error {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return exception.ValidationError{
			Message: "invalid classroom ID format",
		}
	}

	existing, err := s.ClassroomRepository.FindByID(ctx, parsedUUID)
	if err != nil {
		return exception.NotFoundError{
			Message: err.Error(),
		}
	}

	err = s.ClassroomRepository.Delete(ctx, parsedUUID)
	if err != nil {
		return exception.NotFoundError{
			Message: err.Error(),
		}
	}

	// Delete old files from storage on successful deletion
	if existing.CoverImg != nil && *existing.CoverImg != "" {
		_ = s.Storage.DeleteFile(ctx, *existing.CoverImg)
	}
	if existing.WallpaperImg != nil && *existing.WallpaperImg != "" {
		_ = s.Storage.DeleteFile(ctx, *existing.WallpaperImg)
	}

	return nil
}

func (s *classroomServiceImpl) FindAll(ctx context.Context) ([]model.ClassroomResponse, error) {
	classrooms, err := s.ClassroomRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var response []model.ClassroomResponse
	for _, c := range classrooms {
		response = append(response, s.toClassroomResponse(c, true))
	}
	return response, nil
}

func (s *classroomServiceImpl) FindMyClassrooms(ctx context.Context, userID string, isSuperAdmin bool) ([]model.ClassroomResponse, error) {
	parsedUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, exception.ValidationError{
			Message: "invalid user ID format",
		}
	}

	classrooms, err := s.ClassroomRepository.FindAllByUserID(ctx, parsedUUID)
	if err != nil {
		return nil, err
	}

	var roleMap map[uuid.UUID]entity.ClassroomRoleType
	if !isSuperAdmin {
		roles, err := s.ClassroomRepository.FindUserRoles(ctx, parsedUUID)
		if err != nil {
			return nil, err
		}
		roleMap = make(map[uuid.UUID]entity.ClassroomRoleType)
		for _, r := range roles {
			roleMap[r.ClassroomID] = r.Role
		}
	}

	var response []model.ClassroomResponse
	for _, c := range classrooms {
		showSecureFields := isSuperAdmin
		if !isSuperAdmin && roleMap != nil {
			role := roleMap[c.ID]
			showSecureFields = role == entity.RoleOwner || role == entity.RoleTeacher
		}
		response = append(response, s.toClassroomResponse(c, showSecureFields))
	}
	return response, nil
}

func (s *classroomServiceImpl) FindByID(ctx context.Context, id string, userID string, isSuperAdmin bool) (model.ClassroomResponse, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return model.ClassroomResponse{}, exception.ValidationError{
			Message: "invalid classroom ID format",
		}
	}

	classroom, err := s.ClassroomRepository.FindByID(ctx, parsedUUID)
	if err != nil {
		return model.ClassroomResponse{}, exception.NotFoundError{
			Message: err.Error(),
		}
	}

	// Verify membership if not super admin
	showSecureFields := isSuperAdmin
	if !isSuperAdmin {
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			return model.ClassroomResponse{}, exception.ValidationError{
				Message: "invalid user ID format",
			}
		}
		userRole, err := s.ClassroomRepository.FindUserRole(ctx, parsedUUID, userUUID)
		if err != nil {
			return model.ClassroomResponse{}, fiber.NewError(fiber.StatusForbidden, "you are not a member of this classroom")
		}
		showSecureFields = userRole.Role == entity.RoleOwner || userRole.Role == entity.RoleTeacher
	}

	return s.toClassroomResponse(classroom, showSecureFields), nil
}

func (s *classroomServiceImpl) JoinByCode(ctx context.Context, request model.JoinClassroomRequest, userID string) (model.ClassroomResponse, error) {
	common.Validate(request)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return model.ClassroomResponse{}, exception.ValidationError{
			Message: "invalid user ID format",
		}
	}

	classroom, err := s.ClassroomRepository.FindByCode(ctx, request.Code)
	if err != nil {
		return model.ClassroomResponse{}, exception.NotFoundError{
			Message: err.Error(),
		}
	}

	// Check if invite is enabled
	if !classroom.IsExternalInviteEnable {
		return model.ClassroomResponse{}, exception.ValidationError{
			Message: "joining via code is disabled for this classroom",
		}
	}

	// Check if user is already a member
	_, err = s.ClassroomRepository.FindUserRole(ctx, classroom.ID, userUUID)
	if err == nil {
		return model.ClassroomResponse{}, exception.ValidationError{
			Message: "you are already a member of this classroom",
		}
	}

	// Add member as student
	role := entity.ClassroomRole{
		ID:          uuid.New(),
		UserID:      userUUID,
		ClassroomID: classroom.ID,
		Role:        entity.RoleStudent,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = s.ClassroomRepository.AddMember(ctx, role)
	if err != nil {
		return model.ClassroomResponse{}, err
	}

	return s.toClassroomResponse(classroom, false), nil
}

func (s *classroomServiceImpl) ListMembers(ctx context.Context, classroomID string, userID string, isSuperAdmin bool) ([]model.ClassroomMemberResponse, error) {
	classUUID, err := uuid.Parse(classroomID)
	if err != nil {
		return nil, exception.ValidationError{
			Message: "invalid classroom ID format",
		}
	}

	// Verify membership/admin access
	if !isSuperAdmin {
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			return nil, exception.ValidationError{
				Message: "invalid user ID format",
			}
		}
		_, err = s.ClassroomRepository.FindUserRole(ctx, classUUID, userUUID)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusForbidden, "you are not a member of this classroom")
		}
	}

	members, err := s.ClassroomRepository.FindMembers(ctx, classUUID)
	if err != nil {
		return nil, err
	}

	var response []model.ClassroomMemberResponse
	for _, m := range members {
		var lastName string
		if m.User.LastName != nil {
			lastName = *m.User.LastName
		}
		response = append(response, model.ClassroomMemberResponse{
			UserID:    m.UserID.String(),
			Email:     m.User.Email,
			FirstName: m.User.FirstName,
			LastName:  lastName,
			Role:      string(m.Role),
		})
	}

	return response, nil
}

func (s *classroomServiceImpl) toClassroomResponse(c entity.Classroom, showSecureFields bool) model.ClassroomResponse {
	var codes *string
	var isInviteEnable *bool

	if showSecureFields {
		codesVal := c.Codes
		codes = &codesVal
		isInviteEnableVal := c.IsExternalInviteEnable
		isInviteEnable = &isInviteEnableVal
	}

	return model.ClassroomResponse{
		ID:                     c.ID.String(),
		Name:                   c.Name,
		Description:            c.Description,
		Codes:                  codes,
		CoverImg:               c.CoverImg,
		WallpaperImg:           c.WallpaperImg,
		IsExternalInviteEnable: isInviteEnable,
		Status:                 c.Status,
		CreatedAt:              c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              c.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *classroomServiceImpl) UpdateMemberRole(ctx context.Context, classroomID string, targetUserID string, request model.UpdateMemberRoleRequest, currentUserID string, isSuperAdmin bool) error {
	common.Validate(request)

	parsedClassroomID, err := uuid.Parse(classroomID)
	if err != nil {
		return exception.ValidationError{Message: "invalid classroom ID format"}
	}
	parsedTargetUserID, err := uuid.Parse(targetUserID)
	if err != nil {
		return exception.ValidationError{Message: "invalid target user ID format"}
	}
	parsedCurrentUserID, err := uuid.Parse(currentUserID)
	if err != nil {
		return exception.ValidationError{Message: "invalid current user ID format"}
	}

	_, err = s.ClassroomRepository.FindByID(ctx, parsedClassroomID)
	if err != nil {
		return exception.NotFoundError{Message: "classroom not found"}
	}

	if !isSuperAdmin {
		currentUserRole, err := s.ClassroomRepository.FindUserRole(ctx, parsedClassroomID, parsedCurrentUserID)
		if err != nil || currentUserRole.Role != entity.RoleOwner {
			return exception.ValidationError{Message: "only the classroom owner can change member roles"}
		}
	}

	targetUserRole, err := s.ClassroomRepository.FindUserRole(ctx, parsedClassroomID, parsedTargetUserID)
	if err != nil {
		return exception.NotFoundError{Message: "target user is not a member of this classroom"}
	}
	if targetUserRole.Role == entity.RoleOwner {
		return exception.ValidationError{Message: "cannot change the role of the classroom owner"}
	}

	return s.ClassroomRepository.UpdateMemberRole(ctx, parsedClassroomID, parsedTargetUserID, entity.ClassroomRoleType(request.Role))
}

func (s *classroomServiceImpl) RemoveMember(ctx context.Context, classroomID string, targetUserID string, currentUserID string, isSuperAdmin bool) error {
	parsedClassroomID, err := uuid.Parse(classroomID)
	if err != nil {
		return exception.ValidationError{Message: "invalid classroom ID format"}
	}
	parsedTargetUserID, err := uuid.Parse(targetUserID)
	if err != nil {
		return exception.ValidationError{Message: "invalid target user ID format"}
	}
	parsedCurrentUserID, err := uuid.Parse(currentUserID)
	if err != nil {
		return exception.ValidationError{Message: "invalid current user ID format"}
	}

	_, err = s.ClassroomRepository.FindByID(ctx, parsedClassroomID)
	if err != nil {
		return exception.NotFoundError{Message: "classroom not found"}
	}

	var curRole entity.ClassroomRoleType
	if !isSuperAdmin {
		currentUserRole, err := s.ClassroomRepository.FindUserRole(ctx, parsedClassroomID, parsedCurrentUserID)
		if err != nil || (currentUserRole.Role != entity.RoleOwner && currentUserRole.Role != entity.RoleTeacher) {
			return exception.ValidationError{Message: "only the classroom owner or teacher can kick members"}
		}
		curRole = currentUserRole.Role
	}

	targetUserRole, err := s.ClassroomRepository.FindUserRole(ctx, parsedClassroomID, parsedTargetUserID)
	if err != nil {
		return exception.NotFoundError{Message: "target user is not a member of this classroom"}
	}

	if targetUserRole.Role == entity.RoleOwner {
		return exception.ValidationError{Message: "cannot kick the classroom owner"}
	}

	if !isSuperAdmin && curRole == entity.RoleTeacher && targetUserRole.Role == entity.RoleTeacher {
		return exception.ValidationError{Message: "teachers cannot kick other teachers"}
	}

	return s.ClassroomRepository.RemoveMember(ctx, parsedClassroomID, parsedTargetUserID)
}

func (s *classroomServiceImpl) LeaveClassroom(ctx context.Context, classroomID string, currentUserID string) error {
	parsedClassroomID, err := uuid.Parse(classroomID)
	if err != nil {
		return exception.ValidationError{Message: "invalid classroom ID format"}
	}
	parsedCurrentUserID, err := uuid.Parse(currentUserID)
	if err != nil {
		return exception.ValidationError{Message: "invalid current user ID format"}
	}

	_, err = s.ClassroomRepository.FindByID(ctx, parsedClassroomID)
	if err != nil {
		return exception.NotFoundError{Message: "classroom not found"}
	}

	currentUserRole, err := s.ClassroomRepository.FindUserRole(ctx, parsedClassroomID, parsedCurrentUserID)
	if err != nil {
		return exception.NotFoundError{Message: "you are not a member of this classroom"}
	}

	if currentUserRole.Role == entity.RoleOwner {
		return exception.ValidationError{Message: "the owner cannot leave the classroom. Please delete the classroom instead or transfer ownership"}
	}

	return s.ClassroomRepository.RemoveMember(ctx, parsedClassroomID, parsedCurrentUserID)
}

func (s *classroomServiceImpl) GetLeaderboard(ctx context.Context, classroomID string, topicIDStr string, currentUserID string, isSuperAdmin bool) ([]model.LeaderboardEntry, error) {
	parsedClassroomID, err := uuid.Parse(classroomID)
	if err != nil {
		return nil, exception.ValidationError{Message: "invalid classroom ID format"}
	}
	parsedCurrentUserID, err := uuid.Parse(currentUserID)
	if err != nil {
		return nil, exception.ValidationError{Message: "invalid user ID format"}
	}

	_, err = s.ClassroomRepository.FindByID(ctx, parsedClassroomID)
	if err != nil {
		return nil, exception.NotFoundError{Message: "classroom not found"}
	}

	if !isSuperAdmin {
		_, err = s.ClassroomRepository.FindUserRole(ctx, parsedClassroomID, parsedCurrentUserID)
		if err != nil {
			return nil, exception.ValidationError{Message: "only members of this classroom can view the leaderboard"}
		}
	}

	var topicID *uuid.UUID
	if topicIDStr != "" {
		parsedTopicID, err := uuid.Parse(topicIDStr)
		if err != nil {
			return nil, exception.ValidationError{Message: "invalid topic ID format"}
		}
		topicID = &parsedTopicID
	}

	return s.ClassroomRepository.GetLeaderboard(ctx, parsedClassroomID, topicID)
}
