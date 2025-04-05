package cron

import (
	"bpl/client"
	"bpl/parser"
	"bpl/repository"
	"bpl/service"
	"bpl/utils"
	"context"
	"log"
	"os"
	"sync"
	"time"
)

var ascendancyNodesPoE1 = utils.ToSet([]int{193, 258, 409, 607, 662, 758, 869, 922, 982, 1105, 1675, 1697, 1729, 1731, 1734, 1945, 2060, 2336, 2521, 2598, 2872, 3184, 3554, 3651, 4194, 4242, 4494, 4849, 4917, 5082, 5087, 5415, 5443, 5643, 5819, 5865, 5926, 6028, 6038, 6052, 6064, 6728, 6778, 6982, 7618, 8281, 8419, 8592, 8656, 9014, 9271, 9327, 9971, 10099, 10143, 10238, 10635, 11046, 11412, 11490, 11597, 12146, 12475, 12597, 12738, 12850, 13219, 13374, 13851, 14103, 14156, 14603, 14726, 14870, 14996, 15286, 15550, 15616, 16023, 16093, 16212, 16306, 16745, 16848, 16940, 17018, 17315, 17445, 17765, 17988, 18309, 18378, 18574, 18635, 19083, 19417, 19488, 19587, 19595, 19598, 19641, 20050, 20480, 20954, 21264, 22551, 22637, 22852, 23024, 23169, 23225, 23509, 23572, 23972, 24214, 24432, 24528, 24538, 24704, 24755, 24848, 24984, 25111, 25167, 25309, 25651, 26067, 26298, 26446, 26714, 27038, 27055, 27096, 27536, 27604, 27864, 28535, 28782, 28884, 28995, 29026, 29294, 29630, 29662, 29825, 29994, 30690, 30919, 30940, 31316, 31344, 31364, 31598, 31667, 31700, 31984, 32115, 32249, 32251, 32364, 32417, 32640, 32662, 32730, 32816, 32947, 32992, 33167, 33179, 33645, 33795, 33875, 33940, 33954, 34215, 34434, 34484, 34567, 34774, 35185, 35598, 35750, 35754, 36017, 36242, 36958, 37114, 37127, 37191, 37419, 37486, 37492, 37623, 38180, 38387, 38689, 38918, 38999, 39598, 39728, 39790, 39818, 39834, 40010, 40059, 40104, 40510, 40631, 40810, 40813, 41081, 41433, 41534, 41891, 41996, 42144, 42264, 42293, 42546, 42659, 42671, 42861, 43122, 43193, 43195, 43215, 43242, 43336, 43725, 43962, 44297, 44482, 44797, 45313, 45403, 45696, 46952, 47366, 47486, 47630, 47778, 47873, 48124, 48214, 48239, 48480, 48719, 48760, 48904, 48999, 49153, 50024, 50692, 50845, 51101, 51462, 51492, 51998, 52575, 53086, 53095, 53123, 53421, 53816, 53884, 53992, 54159, 54279, 54877, 55146, 55236, 55509, 55646, 55686, 55867, 55985, 56134, 56461, 56722, 56789, 56856, 56967, 57052, 57197, 57222, 57429, 57560, 58029, 58229, 58427, 58454, 58650, 58827, 58998, 59800, 59837, 59920, 60462, 60508, 60547, 60769, 60791, 61072, 61259, 61355, 61372, 61393, 61478, 61627, 61761, 61805, 61871, 62067, 62136, 62162, 62349, 62504, 62595, 62817, 63135, 63293, 63357, 63417, 63490, 63583, 63673, 63908, 63940, 64028, 64111, 64768, 64842, 65153, 65296})
var ascendancyNodesPoE2 = utils.ToSet([]int{16, 30, 40, 59, 74, 110, 528, 664, 762, 770, 1347, 1442, 1579, 1583, 1988, 1994, 2516, 2702, 2857, 2877, 2995, 3065, 3084, 3165, 3704, 3762, 3781, 3987, 4245, 4495, 4891, 5386, 5563, 5817, 5852, 6109, 6127, 6935, 7120, 7246, 7621, 7656, 7793, 7979, 7998, 8143, 8272, 8415, 8525, 8611, 8854, 8867, 9294, 9798, 9988, 9994, 9997, 10072, 10371, 10694, 10731, 10987, 11641, 11771, 11776, 12000, 12054, 12183, 12488, 12795, 12876, 12882, 13065, 13174, 13673, 13675, 13715, 13772, 14429, 14508, 14960, 15044, 16100, 16249, 16276, 16433, 17058, 17268, 17646, 17754, 17788, 17923, 18146, 18158, 18348, 18585, 18678, 18826, 18849, 19233, 19424, 19482, 20195, 20772, 20830, 20895, 22147, 22541, 22661, 22908, 23005, 23352, 23415, 23416, 23508, 23710, 23880, 24039, 24135, 24226, 24295, 24475, 24807, 24868, 25172, 25239, 25434, 25438, 25618, 25779, 25781, 25885, 25935, 26085, 26282, 26638, 27418, 27667, 27686, 27990, 28153, 28431, 29074, 29162, 29323, 29398, 29645, 29871, 30071, 30115, 30151, 30233, 30996, 31116, 31223, 32534, 32559, 32560, 32637, 32699, 32771, 32952, 33141, 33570, 33736, 33812, 34419, 34501, 34817, 34882, 35033, 35187, 35453, 35801, 36252, 36365, 36564, 36659, 36676, 36696, 36728, 36788, 36822, 37046, 37078, 37336, 37397, 37523, 38014, 38578, 38601, 38769, 39204, 39241, 39292, 39365, 39411, 39470, 39640, 39723, 40719, 40721, 40915, 41008, 41076, 41619, 41736, 42017, 42035, 42275, 42416, 42441, 42522, 42845, 43095, 43128, 43131, 44357, 44371, 44484, 44746, 45248, 46016, 46071, 46454, 46522, 46535, 46644, 46990, 47097, 47184, 47236, 47312, 47344, 47442, 48537, 48682, 49049, 49165, 49189, 49340, 49380, 49503, 49759, 50098, 50192, 50219, 51142, 51690, 51737, 52068, 52448, 53108, 53762, 54194, 54838, 54892, 55536, 55582, 55611, 55796, 56162, 56842, 57141, 57181, 57253, 57819, 57959, 58149, 58574, 58591, 58704, 58747, 58751, 58932, 59342, 59372, 59540, 59759, 59822, 59913, 60287, 60298, 60634, 60662, 60859, 60913, 61039, 61267, 61461, 61804, 61897, 61973, 61985, 61991, 62388, 62797, 62804, 63002, 63236, 63254, 63259, 63401, 63484, 63713, 63894, 64031, 64117, 64379, 64789, 64962, 65173, 65413, 65518})

type PlayerFetchingService struct {
	userRepository        *repository.UserRepository
	objectiveMatchService *service.ObjectiveMatchService
	objectiveService      *service.ObjectiveService
	characterService      *service.CharacterService
	ladderService         *service.LadderService
	client                *client.PoEClient
	event                 *repository.Event
}

func NewPlayerFetchingService(client *client.PoEClient, event *repository.Event) *PlayerFetchingService {
	return &PlayerFetchingService{
		userRepository:        repository.NewUserRepository(),
		objectiveMatchService: service.NewObjectiveMatchService(),
		objectiveService:      service.NewObjectiveService(),
		ladderService:         service.NewLadderService(),
		characterService:      service.NewCharacterService(),
		client:                client,
		event:                 event,
	}
}

func (s *PlayerFetchingService) UpdateCharacterName(playerUpdate *parser.PlayerUpdate, event *repository.Event) {
	playerUpdate.Mu.Lock()
	defer playerUpdate.Mu.Unlock()
	if !playerUpdate.ShouldUpdateCharacterName() {
		return
	}
	charactersResponse, err := s.client.ListCharacters(playerUpdate.Token, event.GetRealm())
	playerUpdate.LastUpdateTimes.CharacterName = time.Now()
	if err != nil {
		if err.StatusCode == 401 || err.StatusCode == 403 {
			playerUpdate.TokenExpiry = time.Now()
			return
		}
		log.Print(err)
		return
	}
	for _, char := range charactersResponse.Characters {
		if char.League != nil && *char.League == s.event.Name && char.Level > playerUpdate.New.CharacterLevel {
			playerUpdate.New.CharacterName = char.Name
			playerUpdate.New.CharacterLevel = char.Level
			playerUpdate.New.Ascendancy = char.Class
		}
	}
	log.Printf("Player %s updated: %s (%d)", playerUpdate.AccountName, playerUpdate.New.CharacterName, playerUpdate.New.CharacterLevel)
}

func getAscendancyPoints(character *client.Character, gameVersion repository.GameVersion) int {
	if gameVersion == repository.PoE2 {
		return len(ascendancyNodesPoE2.Intersection(utils.ToSet(character.Passives.Hashes)))
	}
	return len(ascendancyNodesPoE1.Intersection(utils.ToSet(character.Passives.Hashes)))
}

func (s *PlayerFetchingService) UpdateCharacter(player *parser.PlayerUpdate, event *repository.Event) {
	player.Mu.Lock()
	defer player.Mu.Unlock()
	if !player.ShouldUpdateCharacter() {
		return
	}
	characterResponse, err := s.client.GetCharacter(player.Token, player.New.CharacterName, event.GetRealm())
	player.LastUpdateTimes.Character = time.Now()
	if err != nil {
		if err.StatusCode == 401 || err.StatusCode == 403 {
			player.TokenExpiry = time.Now()
			return
		}
		if err.StatusCode == 404 {
			player.New.CharacterName = ""
			return
		}
		log.Print(err)
		return
	}

	player.New.CharacterLevel = characterResponse.Character.Level
	player.New.Pantheon = characterResponse.Character.Passives.PantheonMajor != nil && characterResponse.Character.Passives.PantheonMinor != nil
	player.New.Ascendancy = characterResponse.Character.Class
	player.New.AscendancyPoints = getAscendancyPoints(characterResponse.Character, event.GameVersion)
	player.New.MainSkill = getMainSkill(characterResponse.Character, event.GameVersion)
}

func (s *PlayerFetchingService) UpdateLeagueAccount(player *parser.PlayerUpdate) {
	if s.event.GameVersion == repository.PoE2 {
		return
	}
	player.Mu.Lock()
	defer player.Mu.Unlock()
	if !player.ShouldUpdateLeagueAccount() {
		return
	}
	leagueAccount, err := s.client.GetLeagueAccount(player.Token, s.event.Name)
	player.LastUpdateTimes.LeagueAccount = time.Now()
	if err != nil {
		if err.StatusCode == 401 || err.StatusCode == 403 {
			player.TokenExpiry = time.Now()
			return
		}
		log.Print(err)
		return
	}
	player.New.AtlasPassiveTrees = leagueAccount.LeagueAccount.AtlasPassiveTrees
}

func (s *PlayerFetchingService) UpdateLadder(players []*parser.PlayerUpdate) {
	var resp *client.GetLeagueLadderResponse
	var clientError *client.ClientError
	if s.event.GameVersion == repository.PoE2 {
		// todo: get the ladder for the correct event
		resp, clientError = s.client.GetPoE2Ladder(s.event.Name)
	} else {
		// todo: once we have a token that allows us to request the ladder api
		return
		token := os.Getenv("POE_CLIENT_TOKEN")
		resp, clientError = s.client.GetFullLadder(token, s.event.Name)
	}
	if clientError != nil {
		log.Print(clientError)
		return
	}

	charToUpdate := map[string]*parser.PlayerUpdate{}
	charToUserId := map[string]int{}
	for _, player := range players {
		charToUpdate[player.New.CharacterName] = player
		charToUserId[player.New.CharacterName] = player.UserId
	}

	entriesToPersist := make([]*client.LadderEntry, 0, len(resp.Ladder.Entries))
	for _, entry := range resp.Ladder.Entries {
		if player, ok := charToUpdate[entry.Character.Name]; ok {
			entriesToPersist = append(entriesToPersist, &entry)
			player.Mu.Lock()
			player.New.CharacterLevel = entry.Character.Level
			if entry.Character.Depth != nil && entry.Character.Depth.Depth != nil {
				player.New.DelveDepth = *entry.Character.Depth.Depth
			}
			player.Mu.Unlock()
		}
	}
	err := s.ladderService.UpsertLadder(entriesToPersist, s.event.Id, charToUserId)
	if err != nil {
		log.Print(clientError)
	}
}

func PlayerFetchLoop(ctx context.Context, event *repository.Event, poeClient *client.PoEClient) {
	service := NewPlayerFetchingService(poeClient, event)
	users, err := service.userRepository.GetAuthenticatedUsersForEvent(service.event.Id)
	if err != nil {
		log.Print(err)
		return
	}
	players := utils.Map(users, func(user *repository.TeamUserWithPoEToken) *parser.PlayerUpdate {
		return &parser.PlayerUpdate{
			UserId:      user.UserId,
			TeamId:      user.TeamId,
			AccountName: user.AccountName,
			Token:       user.Token,
			TokenExpiry: user.TokenExpiry,
			New:         parser.Player{},
			Old:         parser.Player{},
			Mu:          sync.Mutex{},
			LastUpdateTimes: struct {
				CharacterName time.Time
				Character     time.Time
				LeagueAccount time.Time
			}{},
		}
	})
	objectives, err := service.objectiveService.GetObjectivesByEventId(service.event.Id)
	if err != nil {
		log.Print(err)
		return
	}
	playerChecker, err := parser.NewPlayerChecker(objectives)
	if err != nil {
		log.Print(err)
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			wg := sync.WaitGroup{}
			// handle character name updates first
			for _, player := range players {
				if player.TokenExpiry.Before(time.Now()) {
					continue
				}
				wg.Add(1)
				go func(player *parser.PlayerUpdate) {
					defer wg.Done()
					service.UpdateCharacterName(player, event)
				}(player)
			}
			wg.Wait()
			wg = sync.WaitGroup{}
			for _, player := range players {
				if player.TokenExpiry.Before(time.Now()) {
					continue
				}
				wg.Add(2)
				go func(player *parser.PlayerUpdate) {
					defer wg.Done()
					service.UpdateCharacter(player, event)
				}(player)
				go func(player *parser.PlayerUpdate) {
					defer wg.Done()
					service.UpdateLeagueAccount(player)
				}(player)
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				service.UpdateLadder(players)
			}()
			wg.Wait()

			for _, player := range players {
				err := service.characterService.SavePlayerUpdate(event.Id, player)
				if err != nil {
					log.Print(err)
				}
			}

			matches := utils.FlatMap(players, func(player *parser.PlayerUpdate) []*repository.ObjectiveMatch {
				return service.GetPlayerMatches(player, playerChecker)
			})
			service.objectiveMatchService.SaveMatches(matches, []int{})
			time.Sleep(1 * time.Minute)
			for _, player := range players {
				player.Old = player.New
			}

		}
	}
}

func (m *PlayerFetchingService) GetPlayerMatches(player *parser.PlayerUpdate, playerChecker *parser.PlayerChecker) []*repository.ObjectiveMatch {
	return utils.Map(playerChecker.CheckForCompletions(player), func(result *parser.CheckResult) *repository.ObjectiveMatch {
		return &repository.ObjectiveMatch{
			ObjectiveId: result.ObjectiveId,
			UserId:      player.UserId,
			Number:      result.Number,
			Timestamp:   time.Now(),
			EventId:     m.event.Id,
		}
	})
}

func getMainSkill(character *client.Character, gameVersion repository.GameVersion) string {
	if gameVersion == repository.PoE2 {
		return getMainSkillPoe2(character)
	}
	return getMainSkillPoe1(character)
}

func getMainSkillPoe1(character *client.Character) string {
	mainSkill := ""
	maxLinks := 0
	for _, item := range *character.Equipment {
		if item.SocketedItems == nil || item.Sockets == nil {
			continue
		}
		socketedItems := *item.SocketedItems
		sockets := *item.Sockets
		for _, gem := range socketedItems {

			if gem.Support == nil || *gem.Support {
				continue
			}
			group := sockets[*gem.Socket].Group
			links := 1
			for socketId, socket := range sockets {
				if len(socketedItems) <= socketId || socket.Group != group {
					continue
				}
				support := socketedItems[socketId].Support
				if support != nil && *support {
					links++
				}
			}
			if links > maxLinks {
				maxLinks = links
				mainSkill = gem.BaseType
			}

		}
	}
	return mainSkill
}
func getMainSkillPoe2(character *client.Character) string {
	mainSkill := ""
	maxLinks := 0
	for _, skillGem := range *character.Skills {
		links := 0
		if (skillGem.Support == nil || *skillGem.Support) || skillGem.Sockets == nil {
			continue
		}
		for _, socket := range *skillGem.Sockets {
			if socket.Item != nil && *socket.Item == client.ItemSocketItemSupportGem {
				links++
			}
		}
		if links > maxLinks {
			maxLinks = links
			mainSkill = skillGem.BaseType
		}
	}
	return mainSkill
}
