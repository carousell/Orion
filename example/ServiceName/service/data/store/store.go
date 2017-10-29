package store

import (
	"context"
	"strings"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/carousell/DataAccessLayer/dal"
	"github.com/carousell/DataAccessLayer/dal/cassandra"
	"github.com/carousell/DataAccessLayer/dal/core"
	"github.com/carousell/DataAccessLayer/dal/es"
	"github.com/carousell/Orion/example/ServiceName/service/data"
	"github.com/carousell/go-utils/utils/errors"
	uuid "github.com/satori/go.uuid"
)

func NewClient(c cassandra.Config, e es.Config) (data.StorageService, error) {
	s := new(storage)
	var err error
	s.dalClient, err = core.NewClient(c, e)
	if err != nil {
		return nil, err
	}
	return s, nil
}

type storage struct {
	dalClient dal.DataAccessLayer
}

func (s *storage) Close() {
	s.dalClient.Close()
}

func (s *storage) AddComment(ctx context.Context, comment string) (string, error) {
	if strings.TrimSpace(comment) == "" {
		return "", errors.New("Comment should not be empty")
	}
	u := uuid.NewV4().String()
	m := data.Message{}
	m.Msg.Scan(comment)
	m.Time.Scan(time.Now())
	m.UUID.Scan(u)
	err := s.dalClient.Insert(ctx, m)
	if err != nil {
		return "", errors.Wrap(err, "error inserting in DAL")
	}
	return u, nil
}

func (s *storage) GetComment(ctx context.Context, UUID string) (string, error) {
	if strings.TrimSpace(UUID) == "" {
		return "", errors.New("Comment should not be empty")
	}
	m := data.Message{}
	m.UUID.Scan(UUID)
	err := s.dalClient.ReadPrimary(ctx, &m)
	if err != nil {
		return "", errors.Wrap(err, "error reading from DAL")
	}
	return m.Msg.String, nil
}

func (s *storage) SearchComments(ctx context.Context, term string) ([]data.Message, error) {

	term = strings.TrimSpace(term)
	if term == "" {
		return []data.Message{}, errors.New("Empty query")
	}

	q := elastic.NewBoolQuery()
	m := elastic.NewMultiMatchQuery(term, "message")
	m.Type("phrase_prefix")
	m.Operator("and")
	q.Must(m)

	res, err := s.dalClient.Find(ctx, data.Message{}, q, 0, 10, []elastic.Sorter{})

	if err != nil || res == nil {
		return []data.Message{}, err
	}
	result := make([]data.Message, 0, len(res))
	for _, item := range res {
		if i, ok := item.(data.Message); ok {
			result = append(result, i)
		}
	}
	return result, err
}
