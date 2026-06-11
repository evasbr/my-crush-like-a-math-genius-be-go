package service

import (
	"context"
	"evasbr/mclamg/model"
	"mime/multipart"
)

type ClassroomService interface {
	Create(ctx context.Context, request model.CreateClassroomRequest, coverHeader *multipart.FileHeader, wallpaperHeader *multipart.FileHeader, creatorID string) (model.ClassroomResponse, error)
	Update(ctx context.Context, request model.UpdateClassroomRequest, coverHeader *multipart.FileHeader, wallpaperHeader *multipart.FileHeader, id string) (model.ClassroomResponse, error)
	Delete(ctx context.Context, id string) error
	FindAll(ctx context.Context) ([]model.ClassroomResponse, error)
	FindMyClassrooms(ctx context.Context, userID string, isSuperAdmin bool) ([]model.ClassroomResponse, error)
	FindByID(ctx context.Context, id string, userID string, isSuperAdmin bool) (model.ClassroomResponse, error)
	JoinByCode(ctx context.Context, request model.JoinClassroomRequest, userID string) (model.ClassroomResponse, error)
	ListMembers(ctx context.Context, classroomID string, userID string, isSuperAdmin bool) ([]model.ClassroomMemberResponse, error)
}
