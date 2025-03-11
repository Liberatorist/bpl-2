package parser

import (
	"bpl/client"
	"bpl/repository"
	"bpl/utils"
	"fmt"
	"sync"
	"time"
)

type Player struct {
	CharacterName     string
	CharacterLevel    int
	Pantheon          bool
	AscendancyPoints  int
	AtlasPassiveTrees []client.AtlasPassiveTree
	DelveDepth        int
}
type PlayerUpdate struct {
	UserId      int
	AccountName string
	TeamId      int
	Token       string
	TokenExpiry time.Time
	Mu          sync.Mutex

	New *Player
	Old *Player

	LastUpdateTimes struct {
		CharacterName time.Time
		Character     time.Time
		LeagueAccount time.Time
	}
}

func (p *Player) maxAtlasTreeNodes() int {
	return utils.Max(utils.Map(p.AtlasPassiveTrees, func(tree client.AtlasPassiveTree) int {
		return len(tree.Hashes)
	}))
}

func (p *PlayerUpdate) ShouldUpdateCharacterName() bool {
	if p.New.CharacterName == "" {
		return time.Since(p.LastUpdateTimes.CharacterName) > 1*time.Minute
	}
	return time.Since(p.LastUpdateTimes.CharacterName) > 10*time.Minute
}

func (p *PlayerUpdate) ShouldUpdateCharacter() bool {
	if p.New.CharacterName == "" {
		return false
	}
	if p.New.CharacterLevel > 40 && !p.New.Pantheon {
		return time.Since(p.LastUpdateTimes.Character) > 1*time.Minute
	}
	if p.New.CharacterLevel > 68 && !(p.New.AscendancyPoints >= 8) {
		return time.Since(p.LastUpdateTimes.Character) > 1*time.Minute
	}
	return time.Since(p.LastUpdateTimes.Character) > 10*time.Minute
}

func (p *PlayerUpdate) ShouldUpdateLeagueAccount() bool {
	if p.New.CharacterLevel < 55 {
		return false
	}

	if p.New.maxAtlasTreeNodes() < 100 {
		return time.Since(p.LastUpdateTimes.LeagueAccount) > 1*time.Minute
	}

	return time.Since(p.LastUpdateTimes.LeagueAccount) > 10*time.Minute
}

type PlayerObjectiveChecker func(p *Player) int

func GetChecker(objective *repository.Objective) (PlayerObjectiveChecker, error) {
	if objective.ObjectiveType != repository.PLAYER {
		return nil, fmt.Errorf("not a player objective")
	}
	switch objective.NumberField {
	case repository.PLAYER_LEVEL:
		return func(p *Player) int {
			return p.CharacterLevel
		}, nil
	case repository.DELVE_DEPTH:
		return func(p *Player) int {
			return p.DelveDepth
		}, nil
	case repository.PANTHEON:
		return func(p *Player) int {
			if p.Pantheon {
				return 1
			}
			return 0
		}, nil
	case repository.ASCENDANCY:
		return func(p *Player) int {
			return p.AscendancyPoints
		}, nil
	case repository.PLAYER_SCORE:
		return func(p *Player) int {
			score := 0
			if p.CharacterLevel >= 75 {
				score += 3
				if p.CharacterLevel >= 90 {
					score += 3
				}
			}

			if p.maxAtlasTreeNodes() > 100 {
				score += 3
			}
			if score > 9 {
				return 9
			}
			return score
		}, nil

	default:
		return nil, fmt.Errorf("unsupported number field")
	}
}

type PlayerChecker map[int]PlayerObjectiveChecker

func NewPlayerChecker(objectives []*repository.Objective) (*PlayerChecker, error) {
	checkers := make(map[int]PlayerObjectiveChecker)
	for _, objective := range objectives {
		if objective.ObjectiveType != repository.PLAYER {
			continue
		}
		checker, err := GetChecker(objective)
		if err != nil {
			return nil, err
		}
		checkers[objective.Id] = checker
	}
	return (*PlayerChecker)(&checkers), nil
}

func (pc *PlayerChecker) CheckForCompletions(update *PlayerUpdate) []*CheckResult {
	results := make([]*CheckResult, 0)
	for id, checker := range *pc {
		new := checker(update.New)
		if new != checker(update.Old) {
			results = append(results, &CheckResult{
				ObjectiveId: id,
				Number:      new,
			})
		}
	}
	return results
}
