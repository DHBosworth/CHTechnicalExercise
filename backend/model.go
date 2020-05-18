package backend

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Game is a container for game data.
// it also implements json.Marshaller so it can easily be encoded into json.
type Game struct {
	Title       string    `json:"title" bson:"title"`
	Description string    `json:"description"`
	By          string    `json:"by"`
	Platform    []string  `json:"platform"`
	AgeRating   string    `json:"age_rating" bson:"age_rating"`
	Likes       int       `json:"likes"`
	Comments    []Comment `json:"comments"`
}

// Comment is a container for comment data.
// It implements json.Marshaller so it can easily be encoded into json.
type Comment struct {
	User        string          `json:"user"`
	Message     string          `json:"message"`
	DateCreated EpochToReadable `json:"dateCreated,string"`
	Like        int             `json:"like"`
}

// Report is a conatiner for report data.
type Report struct {
	UserWithMostComments string             `json:"user_with_most_comments"`
	HighestRatedGame     string             `json:"highest_rated_game"`
	AverageLikesPerGame  []GameAverageLikes `json:"average_likes_per_game"`
}

// GameAverageLikes holds data for the average likes for a specific game.
type GameAverageLikes struct {
	Title        string `json:"title"`
	AverageLikes int    `json:"average_likes"`
}

// Error is the Services error response format.
type Error struct {
	Msg string `json:"error"`
}

// The time format used to format the date
const timeFormat = "2006-01-02"

// EpochToReadable is a type alias used to implement a custom JSON marshaller
// for a human readbale date.
type EpochToReadable time.Time

// Format calls the underlying time.Time's Format method
func (jt EpochToReadable) Format(template string) string {
	return time.Time(jt).Format(template)
}

// MarshalJSON makes EpochToReadable implement json.Marshaller so that it will
// be encoded into a human readable format.
func (jt EpochToReadable) MarshalJSON() ([]byte, error) {
	timeStr := fmt.Sprintf(`"%s"`, jt.Format(timeFormat))
	return []byte(timeStr), nil
}

// UnmarshalJSON makes EpochToReadable implement json.UnMarshaller so that it
// can decode human readable formatted time.
func (jt *EpochToReadable) UnmarshalJSON(data []byte) error {
	timeStr := strings.Trim(string(data), `"`)
	t, err := time.Parse(timeFormat, timeStr)
	if err != nil {
		return fmt.Errorf("Error decoding json time: %v", err)
	}

	*jt = EpochToReadable(t)

	return nil
}

// UnmarshalBSONValue handles the conversion from timestamp to golangs time.Time
// when decoding from mongoDB
func (jt *EpochToReadable) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	v := bson.RawValue{
		Type:  bsontype.Int64,
		Value: data,
	}

	epochTimeStamp, ok := v.Int64OK()
	if !ok {
		log.WithFields(log.Fields{
			"Type": t,
			"Data": data,
		}).Warnf("Error decoding timestamp from database.")

		return nil
	}

	timeValue := time.Unix(epochTimeStamp, 0)
	*jt = EpochToReadable(timeValue)

	return nil
}

// User is a container for data stored about a user
type User struct {
	_id      primitive.ObjectID
	Name     string               `bson:"name"`
	Comments []primitive.ObjectID `bson:"comments"`
}
