package verifythumpysoloqueue

import (
    "discord-bot-go/internal/infra/riot"
    "discord-bot-go/internal/util"
    "errors"
    "fmt"
    "net/http"
    "time"
    "github.com/bwmarrin/discordgo"
)

func VerifyThumpySoloQueueCommand(s *discordgo.Session, m *discordgo.MessageCreate, client *riot.RiotClient) {
    gameName := "CHRIST IS BACK"
    tagLine := "GPTHN"
    
    account, err := client.GetAccountByRiotID(gameName, tagLine)
    if err != nil {
        fmt.Println("Erro ao buscar conta pelo Riot ID:", err)
        return
    }
    
    summoner, err := client.GetSummonerByPUUID(account.Puuid)
    if err != nil {
        fmt.Println("Erro ao buscar summoner:", err)
        return
    }
    
    thumpy_puuid := summoner.Puuid
    
    matchIDs, error_match_ids := client.GetMatchIDs(thumpy_puuid, 1, "", "420")
    if error_match_ids != nil {
        fmt.Println("Erro ao buscar partidas:", error_match_ids)
        return
    }
    
    if len(matchIDs) == 0 {
        fmt.Println("Nenhuma partida encontrada")
        return
    }
    
    match, error_details := findMatchDetails(client, matchIDs[len(matchIDs)-1], 3)
    if error_details != nil {
        fmt.Println("Erro ao buscar detalhes da partida:", error_details)
        return
    }
    
    thumpy, error_thumpy := client.GetAccountInfoByPuuid(thumpy_puuid)
    if error_thumpy != nil {
        fmt.Println("Erro ao buscar detalhes do thumpy:", error_thumpy)
        return
    }
    
    if len(thumpy) == 0 {
        fmt.Println("Thumpy não tem dados de ranked")
        return
    }
    
    var thumpy_solo_duo *riot.RankedAccountInfoDto
    for _, info := range thumpy {
        if info.QueueType == "RANKED_SOLO_5x5" {
            thumpy_solo_duo = info
            break
        }
    }
    
    if thumpy_solo_duo == nil {
        fmt.Println("Thumpy não tem dados de Solo/Duo queue")
        return
    }
    
    if didBroWin(match, thumpy_puuid) {
        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Thumpy ganhou sua última solo duo e está com %d PDL", thumpy_solo_duo.LeaguePoints))
    } else {
        if thumpy_solo_duo.Tier != "MASTER" {
            s.ChannelMessageSend(m.ChannelID, "Thumpy perdeu sua última solo duo e dropou do mestre!")
        } else {
            s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Thumpy perdeu sua última solo duo e está com %d PDL", thumpy_solo_duo.LeaguePoints))
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