package service

import (
	"context"
	"time"

	"connectrpc.com/connect"
	activityv1 "dailyApp/backend/gen/activity/v1"
	"dailyApp/backend/gen/activity/v1/activityv1connect"
	"dailyApp/backend/internal/model"
	"gorm.io/gorm"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ActivityService struct {
	db *gorm.DB
}

func NewActivityService(db *gorm.DB) activityv1connect.ActivityServiceHandler {
	return &ActivityService{db: db}
}

func toActivityProto(a model.Activity) *activityv1.Activity {
	return &activityv1.Activity{
		Id:        a.ID,
		Name:      a.Name,
		ColorCode: a.ColorCode,
		CreatedAt: timestamppb.New(a.CreatedAt),
	}
}

func toActivityLogProto(l model.ActivityLog) *activityv1.ActivityLog {
	var endTime *timestamppb.Timestamp
	if l.EndTime != nil {
		endTime = timestamppb.New(*l.EndTime)
	}
	return &activityv1.ActivityLog{
		Id:         l.ID,
		ActivityId: l.ActivityID,
		StartTime:  timestamppb.New(l.StartTime),
		EndTime:    endTime,
		CreatedAt:  timestamppb.New(l.CreatedAt),
		Memo:       l.Memo,
	}
}

func (s *ActivityService) CreateActivity(ctx context.Context, req *connect.Request[activityv1.CreateActivityRequest]) (*connect.Response[activityv1.CreateActivityResponse], error) {
	activity := model.Activity{
		Name:      req.Msg.Name,
		ColorCode: req.Msg.ColorCode,
	}
	if err := s.db.Create(&activity).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&activityv1.CreateActivityResponse{
		Activity: toActivityProto(activity),
	}), nil
}

func (s *ActivityService) ListActivities(ctx context.Context, req *connect.Request[activityv1.ListActivitiesRequest]) (*connect.Response[activityv1.ListActivitiesResponse], error) {
	var activities []model.Activity
	if err := s.db.Order("created_at asc").Find(&activities).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var protoActivities []*activityv1.Activity
	for _, a := range activities {
		protoActivities = append(protoActivities, toActivityProto(a))
	}

	return connect.NewResponse(&activityv1.ListActivitiesResponse{
		Activities: protoActivities,
	}), nil
}

func (s *ActivityService) UpdateActivity(ctx context.Context, req *connect.Request[activityv1.UpdateActivityRequest]) (*connect.Response[activityv1.UpdateActivityResponse], error) {
	var activity model.Activity
	if err := s.db.First(&activity, "id = ?", req.Msg.Id).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	activity.Name = req.Msg.Name
	activity.ColorCode = req.Msg.ColorCode
	if err := s.db.Save(&activity).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&activityv1.UpdateActivityResponse{
		Activity: toActivityProto(activity),
	}), nil
}

func (s *ActivityService) DeleteActivity(ctx context.Context, req *connect.Request[activityv1.DeleteActivityRequest]) (*connect.Response[activityv1.DeleteActivityResponse], error) {
	if err := s.db.Delete(&model.Activity{}, "id = ?", req.Msg.Id).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&activityv1.DeleteActivityResponse{}), nil
}

func (s *ActivityService) RecordActivity(ctx context.Context, req *connect.Request[activityv1.RecordActivityRequest]) (*connect.Response[activityv1.RecordActivityResponse], error) {
	now := time.Now()
	logData := model.ActivityLog{
		ActivityID: req.Msg.ActivityId,
		StartTime:  now,
	}
	if err := s.db.Create(&logData).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&activityv1.RecordActivityResponse{
		Log: toActivityLogProto(logData),
	}), nil
}

func (s *ActivityService) StopActivity(ctx context.Context, req *connect.Request[activityv1.StopActivityRequest]) (*connect.Response[activityv1.StopActivityResponse], error) {
	var logData model.ActivityLog
	err := s.db.Where("activity_id = ? AND end_time IS NULL", req.Msg.ActivityId).
		Order("start_time desc").First(&logData).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	now := time.Now()
	logData.EndTime = &now
	if err := s.db.Save(&logData).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&activityv1.StopActivityResponse{
		Log: toActivityLogProto(logData),
	}), nil
}

func (s *ActivityService) GetDailyLogs(ctx context.Context, req *connect.Request[activityv1.GetDailyLogsRequest]) (*connect.Response[activityv1.GetDailyLogsResponse], error) {
	dateStr := req.Msg.Date
	jst, _ := time.LoadLocation("Asia/Tokyo")
	startDate, err := time.ParseInLocation("2006-01-02", dateStr, jst)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	endDate := startDate.Add(24 * time.Hour)

	var logs []model.ActivityLog
	err = s.db.Where("start_time >= ? AND start_time < ?", startDate, endDate).
		Order("start_time asc").Find(&logs).Error
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var protoLogs []*activityv1.ActivityLog
	for _, l := range logs {
		protoLogs = append(protoLogs, toActivityLogProto(l))
	}

	return connect.NewResponse(&activityv1.GetDailyLogsResponse{
		Logs: protoLogs,
	}), nil
}

func (s *ActivityService) UpdateActivityLog(ctx context.Context, req *connect.Request[activityv1.UpdateActivityLogRequest]) (*connect.Response[activityv1.UpdateActivityLogResponse], error) {
	var logData model.ActivityLog
	if err := s.db.First(&logData, "id = ?", req.Msg.Id).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	logData.Memo = req.Msg.Memo
	if err := s.db.Save(&logData).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&activityv1.UpdateActivityLogResponse{
		Log: toActivityLogProto(logData),
	}), nil
}
