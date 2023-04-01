//go:generate ../../../tools/readme_config_includer/generator
package minecraft

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Client is a client for the Minecraft server.
type Client interface {
	// Connect establishes a connection to the server.
	Connect() error

	// Players returns the players on the scoreboard.
	Players() ([]string, error)

	// Number of connected players
	PlayerCount() (int, error)

	// Scores return the objective scores for a player.
	Scores(player string) ([]Score, error)
}

// Minecraft is the plugin type.
type Minecraft struct {
	Server   string `toml:"server"`
	Port     string `toml:"port"`
	Password string `toml:"password"`

	client Client
	Log    telegraf.Logger `toml:"-"`
}

func (*Minecraft) SampleConfig() string {
	return sampleConfig
}

func (s *Minecraft) GenericTags() map[string]string {
	return map[string]string{
		"server": s.Server + ":" + s.Port,
		"source": s.Server,
		"port":   s.Port,
	}
}

func (s *Minecraft) Gather(acc telegraf.Accumulator) error {
	if s.client == nil {
		connector := newConnector(s.Server, s.Port, s.Password)
		s.client = newClient(connector)
	}

	players, err := s.client.Players()
	if err != nil {
		return err
	}

	playersCount, err := s.client.PlayerCount()
	if err != nil {
		return err
	}
	tags := s.GenericTags()
	fields := map[string]interface{}{
		"players_count": playersCount,
	}
	acc.AddFields("minecraft", fields, tags)

	for _, player := range players {
		scores, err := s.client.Scores(player)
		if err != nil {
			return err
		}

		tags := s.GenericTags()
		tags["player"] = player

		var fields = make(map[string]interface{}, len(scores))
		for _, score := range scores {
			fields[score.Name] = score.Value
		}

		acc.AddFields("minecraft", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("minecraft", func() telegraf.Input {
		return &Minecraft{
			Server: "localhost",
			Port:   "25575",
		}
	})
}
