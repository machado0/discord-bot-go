package riot

import (
	"discord-bot-go/internal/util"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type RiotClient struct {
	apiKey string
	client *http.Client
}

type MatchResponse struct {
	Info struct {
		QueueID      int    `json:"queueId"`
		QueueName    string `json:"-"`
		Participants []struct {
			Puuid          string `json:"puuid"`
			Win            bool   `json:"win"`
			Champion       string `json:"championName"`
			RiotIdGameName string `json:"riotIdGameName"`
		} `json:"participants"`
	} `json:"info"`
}

type RankedAccountInfoDto struct {
    LeagueID     string `json:"leagueId"`
    QueueType    string `json:"queueType"`
    Tier         string `json:"tier"`
    Rank         string `json:"rank"`
    SummonerID   string `json:"summonerId"`
    LeaguePoints int    `json:"leaguePoints"`
    Wins         int    `json:"wins"`
    Losses       int    `json:"losses"`
    Veteran      bool   `json:"veteran"`
    Inactive     bool   `json:"inactive"`
    FreshBlood   bool   `json:"freshBlood"`
    HotStreak    bool   `json:"hotStreak"`
}

type AccountDto struct {
	Puuid    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

type SummonerDto struct {
    Puuid string `json:"puuid"`
    Name  string `json:"name"`
    Id    string `json:"id"`
    Level int    `json:"summonerLevel"`
}

func NewRiotClient(apiKey string) *RiotClient {
	return &RiotClient{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (r *RiotClient) GetAccountByRiotID(gameName, tagLine string) (*AccountDto, error) {
    url := fmt.Sprintf("https://americas.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s", gameName, tagLine)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("X-Riot-Token", r.apiKey)
    
    resp, err := r.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("riot API error: %d", resp.StatusCode)
    }
    
    var account AccountDto
    if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
        return nil, err
    }
    
    return &account, nil
}

func (r *RiotClient) GetAccountInfoByPuuid(puuid string) ([]*RankedAccountInfoDto, error) {
    url := fmt.Sprintf("https://br1.api.riotgames.com/lol/league/v4/entries/by-puuid/%s", puuid)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("X-Riot-Token", r.apiKey)
    
    resp, err := r.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("riot API error: %d", resp.StatusCode)
    }
    
    var accountDetails []*RankedAccountInfoDto
    if err := json.NewDecoder(resp.Body).Decode(&accountDetails); err != nil {
        return nil, err
    }
    
    return accountDetails, nil
}

func (r *RiotClient) GetMatchIDs(puuid string, count int, typeFilter, queueFilter string) ([]string, error) {
	url :=
		fmt.Sprintf("https://americas.api.riotgames.com/lol/match/v5/matches/by-puuid/%s/ids?count=%d", puuid, count)
	if typeFilter != "" {
		url += "&type=" + typeFilter
	}
	if queueFilter != "" {
		url += "&queue=" + queueFilter
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Riot-Token", r.apiKey)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("riot API error: %d", resp.StatusCode)
	}
	var matchIDs []string
	if err := json.NewDecoder(resp.Body).Decode(&matchIDs); err != nil {
		return nil, err
	}
	return matchIDs, nil
}

func (r *RiotClient) GetMatchDetail(matchID string) (*MatchResponse, error) {
	url := fmt.Sprintf("https://americas.api.riotgames.com/lol/match/v5/matches/%s", matchID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Riot-Token", r.apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				fmt.Printf("Rate limit para %s. Tempo necessário para a próxima busca: %d segundos\n", matchID, seconds)
				time.Sleep(time.Duration(seconds) * time.Second)
			}
		}

		body, _ := io.ReadAll(resp.Body)
		return nil, &util.HttpError{
			StatusCode: resp.StatusCode,
			Msg:        string(body),
		}
	}
	var match MatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&match); err != nil {
		return nil, err
	}

	q := util.NewQueueIdentifier()
	match.Info.QueueName = q.GetQueueNameByID(match.Info.QueueID)

	return &match, nil
}

func (r *RiotClient) GetSummonerByPUUID(accountPuuid string) (*SummonerDto, error) {
    url := fmt.Sprintf("https://br1.api.riotgames.com/lol/summoner/v4/summoners/by-puuid/%s", accountPuuid)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("X-Riot-Token", r.apiKey)
    
    resp, err := r.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("riot API error: %d", resp.StatusCode)
    }
    
    var summoner SummonerDto
    if err := json.NewDecoder(resp.Body).Decode(&summoner); err != nil {
        return nil, err
    }
    
    return &summoner, nil
}

