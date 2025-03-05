package service

import (
	"bpl/client"
	"bpl/repository"
	"bpl/utils"
	"log"
)

type StreamService struct {
	teamRepository *repository.TeamRepository
	userRepository *repository.UserRepository
	twitchClient   *client.TwitchClient
	oauthService   *OauthService
}

func NewStreamService() *StreamService {
	oauthService := NewOauthService()
	s := &StreamService{
		userRepository: repository.NewUserRepository(),
		teamRepository: repository.NewTeamRepository(),
		oauthService:   oauthService,
	}
	token, err := oauthService.GetApplicationToken("twitch")
	if err != nil {
		log.Fatalf("Failed to get twitch token: %v", err)
		return s
	}
	s.twitchClient = client.NewTwitchClient(*token)
	return s

}

func (e *StreamService) GetStreamsForCurrentEvent() ([]*client.TwitchStream, error) {
	streamers, err := e.userRepository.GetStreamersForCurrentEvent()
	if err != nil {
		return nil, err
	}
	token, err := e.oauthService.GetApplicationToken("twitch")
	if err != nil {
		return nil, err
	}
	e.twitchClient.Token = *token

	userMap := make(map[string]int)
	for _, streamer := range streamers {
		userMap[streamer.TwitchId] = streamer.UserId
	}

	streams, err := e.twitchClient.GetAllStreams(utils.Map(streamers, func(user *repository.Streamer) string {
		return user.TwitchId
	}))
	if err != nil {
		return nil, err
	}
	for _, stream := range streams {
		if userId, ok := userMap[stream.UserId]; ok {
			stream.BackendUserId = userId
		}
	}
	return streams, nil
}
