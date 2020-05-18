package backend

import (
	"math"
)

type reportAccumulator struct {
	users     map[string]int
	mostLiked struct {
		title string
		likes int
	}
	averageLikes []GameAverageLikes
}

func newReportAcc() *reportAccumulator {
	return &reportAccumulator{
		users:        make(map[string]int),
		averageLikes: make([]GameAverageLikes, 0),
	}
}

func (acc *reportAccumulator) report() (report Report) {
	var (
		maxComments = 0
		maxName     = ""
	)

	for name, comments := range acc.users {
		if comments > maxComments {
			maxName = name
		}
	}

	return Report{
		UserWithMostComments: maxName,
		HighestRatedGame:     acc.mostLiked.title,
		AverageLikesPerGame:  acc.averageLikes,
	}
}

func (acc *reportAccumulator) processGame(game Game) {
	avg, total := processLikes(game)
	if total > acc.mostLiked.likes {
		acc.mostLiked.likes = total
		acc.mostLiked.title = game.Title
	}
	acc.averageLikes = append(acc.averageLikes, GameAverageLikes{
		Title:        game.Title,
		AverageLikes: avg,
	})

	for _, comment := range game.Comments {
		// Will default to zero
		numberOfComments := acc.users[comment.User]
		acc.users[comment.User] = numberOfComments + 1
	}
}

func processLikes(game Game) (avg, total int) {
	likeSum := 0
	for _, comment := range game.Comments {
		likeSum += comment.Like
	}

	avgLikes := int(math.Ceil(float64(likeSum) / float64(len(game.Comments))))
	return avgLikes, likeSum
}
