package service

import (
	"bpl/repository"
)

type SignupService struct {
	eventRepository  *repository.EventRepository
	signupRepository *repository.SignupRepository
	teamRepository   *repository.TeamRepository
}

func NewSignupService() *SignupService {
	return &SignupService{
		signupRepository: repository.NewSignupRepository(),
		eventRepository:  repository.NewEventRepository(),
		teamRepository:   repository.NewTeamRepository(),
	}
}

func (r *SignupService) CreateSignup(signup *repository.Signup) (*repository.Signup, error) {
	return r.signupRepository.CreateSignup(signup)
}

func (r *SignupService) RemoveSignup(userId int, eventId int) error {
	return r.signupRepository.RemoveSignup(userId, eventId)
}
func (r *SignupService) GetSignupForUser(userId int, eventId int) (*repository.Signup, error) {
	return r.signupRepository.GetSignupForUser(userId, eventId)
}

type SignupWithUser struct {
	Signup   repository.Signup
	TeamUser *repository.TeamUser
}

func (r *SignupService) GetSignupsForEvent(eventId int) (map[int][]*repository.Signup, error) {

	event, err := r.eventRepository.GetEventById(eventId, "Teams")
	if err != nil {
		return nil, err
	}
	teamUsers, err := r.teamRepository.GetTeamUsersForEvent(event)
	if err != nil {
		return nil, err
	}
	signups, err := r.signupRepository.GetSignupsForEvent(eventId, event.MaxSize)
	if err != nil {
		return nil, err
	}
	userToTeam := make(map[int]int)
	for _, teamUser := range teamUsers {
		userToTeam[teamUser.UserId] = teamUser.TeamId
	}
	teamSignups := make(map[int][]*repository.Signup)
	for _, signup := range signups {
		teamId, ok := userToTeam[signup.UserId]
		if !ok {
			teamId = 0
		}
		teamSignups[teamId] = append(teamSignups[teamId], signup)
	}

	return teamSignups, nil
}
