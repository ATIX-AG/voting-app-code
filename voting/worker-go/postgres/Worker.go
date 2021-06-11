package postgres

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	m "atix.de/voting/worker-go/model"
	pgx "github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

type worker struct {
	con *pgx.Conn
}

type Worker interface {
	Write(<-chan string) error
}

var ctx = context.Background()

const votesKey = "votes"

func NewWorker(url string) worker {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	con, err := pgx.Connect(context.Background(), url)
	if err != nil {
		log.WithField("url", url).
			Error("unable to connect to database")
	}

	// close connection on stop
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		_ = <-sigc
		con.Close(context.Background())
		os.Exit(0)
	}()

	return worker{
		con: con,
	}
}

func (w worker) Write(channel <-chan m.Vote) error {
	e0 := w.createTable()
	if e0 != nil {
		log.Warn(e0)
	}
	for vote := range channel {
		err := w.post(vote)
		if err != nil {
			log.Warn(err)
		}
	}
	return nil
}

func (w worker) createTable() error {
	resp, err := w.con.Query(ctx, `CREATE TABLE IF NOT EXISTS votes (
		id VARCHAR(255) NOT NULL UNIQUE,
		vote VARCHAR(255) NOT NULL
	)`)
	resp.Close()
	return err
}

func (w worker) post(vote m.Vote) error {
	stmt := `
	INSERT INTO votes (id, vote)
	VALUES ($1, $2)
	ON CONFLICT (id)
		DO UPDATE SET vote = $2;
	`

	resp, err := w.con.Query(ctx, stmt, vote.Id, vote.Vote)
	resp.Close()
	return err
}
