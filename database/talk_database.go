package database

import (
	"database/sql"
	"log"

	"sync"

	"errors"

	"github.com/FEUPTalks/Backend/model"

	//loading the driver anonymously, aliasing its package qualifier to so none of its exported names are visible to our code
	"github.com/FEUPTalks/Backend/model/talkState/talkStateFactory"
	_ "github.com/go-sql-driver/mysql"
)

//TalkDatabaseManager used to manage the talk_store
type talkDatabaseManager struct {
	database *sql.DB
}

const (
	driverName       string = "mysql"
	connectionString string = "lesteamb:99RedBalloons@tcp(127.0.0.1:3306)/talk_store?parseTime=true"
)

var instance *talkDatabaseManager
var once sync.Once

//GetTalkDatabaseManagerInstance returns singleton instance
func GetTalkDatabaseManagerInstance() (*talkDatabaseManager, error) {
	once.Do(func() {
		var db *sql.DB
		var err error

		db, err = sql.Open(driverName, connectionString)
		if err != nil {
			db.Close()
			log.Fatal(err)
		}
		instance = &talkDatabaseManager{db}
	})
	if instance != nil {
		return instance, nil
	}
	return nil, errors.New("Unable to create access to the database")
}

func (manager *talkDatabaseManager) CloseConnection() (err error) {
	err = manager.database.Close()
	if err != nil {
		log.Println(err)
	}
	return
}

func (manager *talkDatabaseManager) Ping() error {
	err := manager.database.Ping()
	if err != nil {
		log.Println(err)
		return errors.New("Unable to access database")
	}
	return nil
}

//GetAllTalks retrieves all talks from the database
func (manager *talkDatabaseManager) GetAllTalks() ([]*model.Talk, error) {
	talks := make([]*model.Talk, 0)
	rows, err := manager.database.Query("select * from talk")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var talk = model.NewTalk()
		var stateTemp uint8
		err := rows.Scan(&talk.TalkID, &talk.Title, &talk.Summary,
			&talk.ProposedInitialDate, &talk.ProposedEndDate,
			&talk.DefinitiveDate, &talk.Duration, &talk.ProponentName,
			&talk.ProponentEmail, &talk.ProponentAffiliation, &talk.SpeakerName,
			&talk.SpeakerBrief, &talk.SpeakerAffiliation, &talk.SpeakerPicture, &talk.HostName,
			&talk.HostEmail, &talk.Snack, &talk.Room, &stateTemp)
		if err != nil {
			log.Println(err)
			continue
		}
		tempState, err := talkStateFactory.GetTalkState(stateTemp)
		if err != nil {
			log.Println(err)
			continue
		}
		talk.SetState(tempState)
		talks = append(talks, talk)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return talks, nil
}

//GetTalk retrieves talks with specific id from the database
func (manager *talkDatabaseManager) GetTalk(talkID int) (*model.Talk, error) {
	stmt, err := manager.database.Prepare("select * from talk where talkID = ?")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	var talk = model.NewTalk()
	var stateTemp uint8

	err = stmt.QueryRow(talkID).Scan(&talk.TalkID, &talk.Title, &talk.Summary,
		&talk.ProposedInitialDate, &talk.ProposedEndDate,
		&talk.DefinitiveDate, &talk.Duration, &talk.ProponentName,
		&talk.ProponentEmail, &talk.ProponentAffiliation, &talk.SpeakerName,
		&talk.SpeakerBrief, &talk.SpeakerAffiliation, &talk.HostName,
		&talk.HostEmail, &talk.Snack, &talk.Room, &stateTemp)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tempState, err := talkStateFactory.GetTalkState(stateTemp)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	talk.SetState(tempState)

	return talk, nil
}

func (manager *talkDatabaseManager) SaveTalk(talk *model.Talk) error {
	stmt, err := manager.database.Prepare(`insert into talk (Title, Summary, ProposedInitialDate,
											ProposedEndDate, ProponentName, ProponentEmail, ProponentAffiliation,
											SpeakerName, SpeakerBrief, SpeakerAffiliation, HostName, HostEmail, State)
											values (?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = stmt.Exec(talk.Title, talk.Summary, talk.ProposedInitialDate, talk.ProposedEndDate,
		talk.ProponentName, talk.ProponentEmail, talk.ProponentAffiliation,
		talk.SpeakerName, talk.SpeakerBrief, talk.SpeakerAffiliation,
		talk.HostName, talk.HostEmail, talk.GetStateValue())
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

//SetTalk
func (manager *talkDatabaseManager) SetTalk(talk *model.Talk) error {
	stmt, err := manager.database.Prepare(`
	UPDATE Talk SET 
		Title=?,
		Summary=?,
		ProposedInitialDate=?,
		ProposedEndDate=?,
		DefinitiveDate=?,
		Duration=?,
		ProponentName=?,
		ProponentEmail=?,
		ProponentAffiliation=?,
		SpeakerName=?,
		SpeakerBrief=?,
		SpeakerAffiliation=?,
		SpeakerPicture=?,
		HostName=?,
		HostEmail=?,
		Snack=?,
		Room=?,
		State=?
	WHERE TalkID=?`)

	if err != nil {
		log.Println(err)
		return err
	}

	_, err = stmt.Exec(talk.Title, talk.Summary, talk.ProposedInitialDate, talk.ProposedEndDate, talk.DefinitiveDate,
		talk.Duration,
		talk.ProponentName, talk.ProponentEmail, talk.ProponentAffiliation,
		talk.SpeakerName, talk.SpeakerBrief, talk.SpeakerAffiliation, talk.SpeakerPicture,
		talk.HostName, talk.HostEmail, talk.Snack, talk.Room, talk.GetStateValue(), talk.TalkID)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}