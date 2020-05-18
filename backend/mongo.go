package backend

import (
	"context"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	gameDatabaseName        = "gamesService"
	gameCollectionName      = "denormalisedGames"
	commentsCollectionName  = "comments"
	publisherCollectionName = "publishers"
	userCollectionName      = "users"
)

// MongoDataSource implements backend.ServiceDataSource so that it can be used
// as the backend for the microservice
type MongoDataSource struct {
	client        *mongo.Client
	gamesDatabase *mongo.Database
}

// NewMongoDataSource creates a new mongo data source
func NewMongoDataSource(addr string) (dataSource *MongoDataSource, err error) {
	options := options.Client().ApplyURI(addr)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options)
	if err != nil {
		return dataSource, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return dataSource, err
	}

	return &MongoDataSource{
		client:        client,
		gamesDatabase: client.Database(gameDatabaseName),
	}, err
}

// Game retrieves information for a game with the given id
func (mongo *MongoDataSource) Game(id string) (game Game, err error) {
	gameCollection := mongo.gamesDatabase.Collection(gameCollectionName)
	filter := bson.M{"id": id}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = gameCollection.FindOne(ctx, filter).Decode(&game)
	if err != nil {
		return game, err
	}

	return game, err
}

// Report creates a report from the stored game data
func (mongo *MongoDataSource) Report() (report Report, err error) {
	report.UserWithMostComments, err = mongo.mostCommentedUser()
	if err != nil {
		log.Warnf("Unable to get user with most comments: %v", err)
	}

	err = mongo.gamesReport(&report)
	if err != nil {
		log.Warnf("Unable to analyse likes for report: %v", err)
	}

	return report, nil
}

func (mongo *MongoDataSource) mostCommentedUser() (name string, err error) {
	gamesCollection := mongo.gamesDatabase.Collection(gameCollectionName)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	cur, err := gamesCollection.Aggregate(ctx, commentsPerUserPipeline())
	if err != nil {
		log.Warnf("Error: %v", err)
		return name, err
	}

	var bestUser userResult
	if cur.Next(ctx) {
		err := cur.Decode(&bestUser)
		if err != nil {
			log.Warnf("Error decoding user: %v", err)
		}
	}

	return bestUser.Name, err
}

type userResult struct {
	Name     string `bson:"_id"`
	Comments int    `bson:"number_of_comments"`
}

func (mongo *MongoDataSource) gamesReport(report *Report) error {
	gamesCollection := mongo.gamesDatabase.Collection(gameCollectionName)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	cur, err := gamesCollection.Aggregate(ctx, gameLikePipeline())
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	first := true
	for cur.Next(ctx) {
		var res gameLikeResult
		err = cur.Decode(&res)
		if err != nil {
			return err
		}

		if first {
			report.HighestRatedGame = res.Title
		}

		report.AverageLikesPerGame = append(report.AverageLikesPerGame, GameAverageLikes{
			Title:        res.Title,
			AverageLikes: int(math.Ceil(res.AvgLikes)),
		})
	}

	return nil
}

type gameLikeResult struct {
	Title    string  `bson:"_id"`
	Likes    int     `bson:"likes"`
	AvgLikes float64 `bson:"avg_likes"`
}

func gameLikePipeline() []bson.D {
	getComments := bson.D{
		{"$unwind", bson.D{
			{"path", "$comments"},
			{"includeArrayIndex", "string"},
			{"preserveNullAndEmptyArrays", true},
		}},
	}

	projectComments := bson.D{
		{
			"$project", bson.D{
				{"comment", "$comments"},
				{"title", "$title"},
			},
		},
	}

	groupByTitle := bson.D{
		{
			"$group", bson.D{
				{
					"_id", "$title",
				},
				{
					"number_of_comments", bson.D{{"$sum", 1}},
				},
				{
					"total_likes", bson.D{{"$sum", "$comment.like"}},
				},
			},
		},
	}

	averageProjection := bson.D{
		{
			"$project", bson.D{
				{
					"avg_likes", bson.D{
						{"$divide", bson.A{"$total_likes", "$number_of_comments"}},
					},
				},
				{"likes", "$total_likes"},
			},
		},
	}

	sort := bson.D{
		{"$sort", bson.D{{"likes", -1}}},
	}

	return []bson.D{
		getComments,
		projectComments,
		groupByTitle,
		averageProjection,
		sort,
	}
}

func commentsPerUserPipeline() []bson.D {
	getComments := bson.D{
		{"$unwind", bson.D{
			{"path", "$comments"},
			{"includeArrayIndex", "string"},
			{"preserveNullAndEmptyArrays", true},
		}},
	}

	projectComments := bson.D{
		{
			"$project", bson.D{
				{"comment", "$comments"},
			},
		},
	}

	groupByName := bson.D{
		{
			"$group", bson.D{
				{"_id", "$comment.user"},
				{"number_of_comments", bson.D{
					{"$sum", 1},
				}},
			},
		},
	}

	sort := bson.D{
		{
			"$sort", bson.D{
				{"number_of_comments", -1},
			},
		},
	}

	return []bson.D{
		getComments,
		projectComments,
		groupByName,
		sort,
	}
}
