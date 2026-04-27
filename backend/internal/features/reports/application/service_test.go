package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"pdd-service/internal/core/domain/users"
	"pdd-service/internal/core/domain/violations"
	coreerrors "pdd-service/internal/core/errors"

	"github.com/google/uuid"
)

func TestServiceCreate(t *testing.T) {
	service, repo := newTestService()
	userID := uuid.New()
	violationTypeID := repo.addViolationType()

	output, err := service.Create(context.Background(), createInput(userID, violationTypeID))
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if output.Report.UserID != userID {
		t.Fatalf("Create() user id = %s, want %s", output.Report.UserID, userID)
	}

	input := createInput(userID, uuid.New())
	if _, err := service.Create(context.Background(), input); !errors.Is(err, coreerrors.ErrViolationTypeNotFound) {
		t.Fatalf("Create() missing violation type error = %v, want ErrViolationTypeNotFound", err)
	}

	input = createInput(userID, violationTypeID)
	input.Video = VideoInput{}
	if _, err := service.Create(context.Background(), input); !errors.Is(err, coreerrors.ErrVideoRequired) {
		t.Fatalf("Create() without video error = %v, want ErrVideoRequired", err)
	}
}

func TestServiceListPermissions(t *testing.T) {
	service, repo := newTestService()
	violationTypeID := repo.addViolationType()
	ownerID := uuid.New()
	otherID := uuid.New()
	repo.addReport(t, ownerID, violationTypeID)
	repo.addReport(t, otherID, violationTypeID)

	userList, err := service.List(context.Background(), ListInput{
		UserID:        &otherID,
		CurrentUserID: ownerID,
		CurrentRole:   users.RoleUser,
	})
	if err != nil {
		t.Fatalf("List() user error = %v", err)
	}
	if userList.Total != 1 || userList.Reports[0].UserID != ownerID {
		t.Fatalf("List() user output = %+v", userList)
	}

	moderatorList, err := service.List(context.Background(), ListInput{
		CurrentUserID: uuid.New(),
		CurrentRole:   users.RoleModerator,
	})
	if err != nil {
		t.Fatalf("List() moderator error = %v", err)
	}
	if moderatorList.Total != 2 {
		t.Fatalf("List() moderator total = %d, want 2", moderatorList.Total)
	}
}

func TestServiceGetPermissions(t *testing.T) {
	service, repo := newTestService()
	violationTypeID := repo.addViolationType()
	ownerID := uuid.New()
	report := repo.addReport(t, ownerID, violationTypeID)

	if _, err := service.GetByID(context.Background(), GetByIDInput{
		ID: report.ID, CurrentUserID: ownerID, CurrentRole: users.RoleUser,
	}); err != nil {
		t.Fatalf("GetByID() owner error = %v", err)
	}

	if _, err := service.GetByID(context.Background(), GetByIDInput{
		ID: report.ID, CurrentUserID: uuid.New(), CurrentRole: users.RoleUser,
	}); !errors.Is(err, coreerrors.ErrForbidden) {
		t.Fatalf("GetByID() other user error = %v, want ErrForbidden", err)
	}
}

func TestServiceUpdatePermissions(t *testing.T) {
	service, repo := newTestService()
	violationTypeID := repo.addViolationType()
	ownerID := uuid.New()
	draft := repo.addReport(t, ownerID, violationTypeID)
	title := "Updated title"

	updated, err := service.Update(context.Background(), UpdateInput{
		ID: draft.ID, Title: &title, CurrentUserID: ownerID, CurrentRole: users.RoleUser,
	})
	if err != nil {
		t.Fatalf("Update() draft error = %v", err)
	}
	if updated.Report.Title != title {
		t.Fatalf("Update() title = %q, want %q", updated.Report.Title, title)
	}

	submitted := repo.addReport(t, ownerID, violationTypeID)
	must(t, submitted.Submit())
	repo.reports[submitted.ID] = submitted
	if _, err := service.Update(context.Background(), UpdateInput{
		ID: submitted.ID, Title: &title, CurrentUserID: ownerID, CurrentRole: users.RoleUser,
	}); err != nil {
		t.Fatalf("Update() submitted error = %v", err)
	}

	approved := repo.addReport(t, ownerID, violationTypeID)
	must(t, approved.Submit())
	must(t, approved.StartReview(uuid.New()))
	must(t, approved.Approve(uuid.New(), "ok"))
	repo.reports[approved.ID] = approved
	if _, err := service.Update(context.Background(), UpdateInput{
		ID: approved.ID, Title: &title, CurrentUserID: ownerID, CurrentRole: users.RoleUser,
	}); !errors.Is(err, coreerrors.ErrForbidden) {
		t.Fatalf("Update() approved error = %v, want ErrForbidden", err)
	}
}

func TestServiceDeleteAndWorkflow(t *testing.T) {
	service, repo := newTestService()
	violationTypeID := repo.addViolationType()
	ownerID := uuid.New()

	draft := repo.addReport(t, ownerID, violationTypeID)
	if err := service.Delete(context.Background(), DeleteInput{
		ID: draft.ID, CurrentUserID: ownerID, CurrentRole: users.RoleUser,
	}); err != nil {
		t.Fatalf("Delete() draft error = %v", err)
	}

	submitted := repo.addReport(t, ownerID, violationTypeID)
	must(t, submitted.Submit())
	repo.reports[submitted.ID] = submitted
	if err := service.Delete(context.Background(), DeleteInput{
		ID: submitted.ID, CurrentUserID: ownerID, CurrentRole: users.RoleUser,
	}); !errors.Is(err, coreerrors.ErrForbidden) {
		t.Fatalf("Delete() submitted error = %v, want ErrForbidden", err)
	}

	draft = repo.addReport(t, ownerID, violationTypeID)
	submittedOutput, err := service.Submit(context.Background(), SubmitInput{
		ID: draft.ID, CurrentUserID: ownerID, CurrentRole: users.RoleUser,
	})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if submittedOutput.Report.Status != violations.StatusSubmitted {
		t.Fatalf("Submit() status = %s", submittedOutput.Report.Status)
	}

	moderatorID := uuid.New()
	inReview, err := service.StartReview(context.Background(), StartReviewInput{
		ID: draft.ID, CurrentUserID: moderatorID, CurrentRole: users.RoleModerator,
	})
	if err != nil {
		t.Fatalf("StartReview() error = %v", err)
	}
	if inReview.Report.Status != violations.StatusInReview {
		t.Fatalf("StartReview() status = %s", inReview.Report.Status)
	}

	approved, err := service.Approve(context.Background(), ApproveInput{
		ID: draft.ID, CurrentUserID: moderatorID, CurrentRole: users.RoleModerator, Comment: "ok",
	})
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if approved.Report.Status != violations.StatusApproved {
		t.Fatalf("Approve() status = %s", approved.Report.Status)
	}

	if _, err := service.Reject(context.Background(), RejectInput{
		ID: draft.ID, CurrentUserID: moderatorID, CurrentRole: users.RoleModerator, Comment: "bad",
	}); !errors.Is(err, coreerrors.ErrInvalidReportStatus) {
		t.Fatalf("Reject() approved error = %v, want ErrInvalidReportStatus", err)
	}
}

func TestServiceReject(t *testing.T) {
	service, repo := newTestService()
	violationTypeID := repo.addViolationType()
	ownerID := uuid.New()
	report := repo.addReport(t, ownerID, violationTypeID)
	must(t, report.Submit())
	must(t, report.StartReview(uuid.New()))
	repo.reports[report.ID] = report

	rejected, err := service.Reject(context.Background(), RejectInput{
		ID: report.ID, CurrentUserID: uuid.New(), CurrentRole: users.RoleAdmin, Comment: "bad video",
	})
	if err != nil {
		t.Fatalf("Reject() error = %v", err)
	}
	if rejected.Report.Status != violations.StatusRejected {
		t.Fatalf("Reject() status = %s", rejected.Report.Status)
	}
}

func createInput(userID uuid.UUID, violationTypeID uuid.UUID) CreateInput {
	return CreateInput{
		UserID:          userID,
		ViolationTypeID: violationTypeID,
		Title:           "Broken parking",
		Description:     "Car parked on sidewalk",
		Location:        "Main street",
		OccurredAt:      time.Now().UTC(),
		Video: VideoInput{
			Source: violations.SourceExternalURL,
			URL:    "https://example.com/video.mp4",
		},
		CurrentUserID: userID,
		CurrentRole:   users.RoleUser,
	}
}

func newTestService() (*Service, *reportsRepositoryStub) {
	repo := &reportsRepositoryStub{
		reports:        map[uuid.UUID]violations.Report{},
		violationTypes: map[uuid.UUID]violations.ViolationType{},
	}
	return NewService(repo, repo), repo
}

type reportsRepositoryStub struct {
	reports        map[uuid.UUID]violations.Report
	violationTypes map[uuid.UUID]violations.ViolationType
}

func (r *reportsRepositoryStub) addViolationType() uuid.UUID {
	violationType, _ := violations.NewViolationType("PARKING", "Parking", "", "500")
	r.violationTypes[violationType.ID] = *violationType
	return violationType.ID
}

func (r *reportsRepositoryStub) addReport(t *testing.T, userID uuid.UUID, violationTypeID uuid.UUID) violations.Report {
	t.Helper()

	video, err := violations.NewExternalVideo("https://example.com/video.mp4")
	must(t, err)
	report, err := violations.NewReport(userID, violationTypeID, "Title", "Description", "Location", time.Now().UTC(), video)
	must(t, err)
	r.reports[report.ID] = *report
	return *report
}

func (r *reportsRepositoryStub) CreateReport(_ context.Context, report violations.Report) error {
	r.reports[report.ID] = report
	return nil
}

func (r *reportsRepositoryStub) GetReportByID(_ context.Context, id uuid.UUID) (violations.Report, error) {
	report, ok := r.reports[id]
	if !ok {
		return violations.Report{}, coreerrors.ErrReportNotFound
	}
	return report, nil
}

func (r *reportsRepositoryStub) UpdateReport(_ context.Context, report violations.Report) error {
	if _, ok := r.reports[report.ID]; !ok {
		return coreerrors.ErrReportNotFound
	}
	r.reports[report.ID] = report
	return nil
}

func (r *reportsRepositoryStub) DeleteReport(_ context.Context, id uuid.UUID) error {
	if _, ok := r.reports[id]; !ok {
		return coreerrors.ErrReportNotFound
	}
	delete(r.reports, id)
	return nil
}

func (r *reportsRepositoryStub) ListReports(_ context.Context, filter ListReportsFilter) ([]violations.Report, error) {
	reports := make([]violations.Report, 0, len(r.reports))
	for _, report := range r.reports {
		if filter.UserID != nil && report.UserID != *filter.UserID {
			continue
		}
		reports = append(reports, report)
	}
	return reports, nil
}

func (r *reportsRepositoryStub) CountReports(ctx context.Context, filter ListReportsFilter) (int64, error) {
	reports, err := r.ListReports(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int64(len(reports)), nil
}

func (r *reportsRepositoryStub) ListReportsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]violations.Report, error) {
	return r.ListReports(ctx, ListReportsFilter{UserID: &userID, Limit: limit, Offset: offset})
}

func (r *reportsRepositoryStub) GetViolationTypeByID(_ context.Context, id uuid.UUID) (violations.ViolationType, error) {
	violationType, ok := r.violationTypes[id]
	if !ok {
		return violations.ViolationType{}, coreerrors.ErrViolationTypeNotFound
	}
	return violationType, nil
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
