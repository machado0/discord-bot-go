package verifythumpy

import (
	"discord-bot-go/internal/infra/riot"
	"discord-bot-go/internal/util"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
)

var gameName = "CHRIST IS BACK"
var tagLine = "GPTHN"

func VerifyThumpyCommand(s *discordgo.Session, m *discordgo.MessageCreate, client *riot.RiotClient) {
	playerSoloduo, match, _ := GetThumpyLastSoloDuo(client)

	if playerSoloduo == nil {
		fmt.Println("Thumpy não tem dados de Solo/Duo queue")
		return
	}

	if didBroWin(match, os.Getenv("THUMPY")) {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Thumpy ganhou sua última solo duo e está com %d PDL", playerSoloduo.LeaguePoints))
	} else {
		ranks := []string{"MASTER", "GRANDMASTER", "CHALLENGER"}

		if slices.Contains(ranks, playerSoloduo.Tier) {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Thumpy perdeu sua última solo duo e está com %d PDL", playerSoloduo.LeaguePoints))
		} else {
			s.ChannelMessageSend(m.ChannelID, "Thumpy perdeu sua última solo duo e dropou do mestre!")
		}
	}
}

func didBroWin(match *riot.MatchResponse, puuid string) bool {
	for _, p := range match.Info.Participants {
		if p.Puuid == puuid {
			return p.Win
		}
	}
	fmt.Println("Erro ao ver se thumpy perdeu ou ganhou a ultima soloduo")
	return false
}

func findMatchDetails(client *riot.RiotClient, matchID string, attempts int) (*riot.MatchResponse, error) {
	var err error
	var match *riot.MatchResponse
	for i := 0; i < attempts; i++ {
		match, err = client.GetMatchDetail(matchID)
		if err == nil {
			return match, nil
		}
		var httpErr *util.HttpError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusTooManyRequests {
			wait := time.Duration((i+1)*2) * time.Second
			fmt.Printf("Rate limit para %s. Tentativa %d. Esperando %v...\n", matchID, i+1, wait)
			time.Sleep(wait)
			continue
		}
		break
	}
	return nil, err
}

func GetThumpyLastSoloDuo(client *riot.RiotClient) (*riot.RankedAccountInfoDto, *riot.MatchResponse, string) {

	account, err := client.GetAccountByRiotID(gameName, tagLine)
	if err != nil {
		fmt.Println("Erro ao buscar conta pelo Riot ID:", err)
	}

	summoner, err := client.GetSummonerByPUUID(account.Puuid)
	if err != nil {
		fmt.Println("Erro ao buscar summoner:", err)
	}

	thumpy_puuid := summoner.Puuid

	matchIDs, error_match_ids := client.GetMatchIDs(thumpy_puuid, 1, "", "420")
	if error_match_ids != nil {
		fmt.Println("Erro ao buscar partidas:", error_match_ids)
	}

	if len(matchIDs) == 0 {
		fmt.Println("Nenhuma partida encontrada")
	}

	match, error_details := findMatchDetails(client, matchIDs[len(matchIDs)-1], 3)
	if error_details != nil {
		fmt.Println("Erro ao buscar detalhes da partida:", error_details)
	}

	playerSoloduo := getSoloduoSummoner(client, summoner.Puuid)

	return playerSoloduo, match, matchIDs[len(matchIDs)-1]
}

func SendDiscordMessageAfterMatch(client *riot.RiotClient, s *discordgo.Session, match *riot.MatchResponse, puuid string) {
	playerSoloduo := getSoloduoSummoner(client, puuid)

	channel := os.Getenv("GAPATHON_CHANNEL")

	if didBroWin(match, puuid) {
		s.ChannelMessageSend(channel, fmt.Sprintf("Thumpy ganhou sua última solo duo e está com %d PDL", playerSoloduo.LeaguePoints))
	} else {
		ranks := []string{"MASTER", "GRANDMASTER", "CHALLENGER"}

		if slices.Contains(ranks, playerSoloduo.Tier) {
			s.ChannelMessageSend(channel, fmt.Sprintf("Thumpy perdeu sua última solo duo e está com %d PDL", playerSoloduo.LeaguePoints))
		} else {
			s.ChannelMessageSend(channel, "Thumpy perdeu sua última solo duo e dropou do mestre!")
		}
	}
}

func getSoloduoSummoner(client *riot.RiotClient, puuid string) *riot.RankedAccountInfoDto {
	player, error_player := client.GetAccountInfoByPuuid(puuid)
	if error_player != nil {
		fmt.Println("Erro ao buscar detalhes do thumpy:", error_player)
	}

	if len(player) == 0 {
		fmt.Println("Thumpy não tem dados de ranked")
	}

	var player_solo_duo *riot.RankedAccountInfoDto
	for _, info := range player {
		if info.QueueType == "RANKED_SOLO_5x5" {
			player_solo_duo = info
			break
		}
	}

	return player_solo_duo
}
