package arsha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PostId struct {
	Id int64 `json:"id"`
}

type BiddingInfo struct {
	Name   string `json:"name"`
	Id     int64  `json:"id"`
	Sid    int64  `json:"sid"`
	Orders []struct {
		Price   int64 `json:"price"`
		Sellers int64 `json:"sellers"`
		Buyers  int64 `json:"buyers"`
	} `json:"orders"`
}

type MarketItem struct {
	Name         string `json:"name"`
	Id           int64  `json:"id"`
	CurrentStock int64  `json:"currentStock"`
	TotalTrades  int64  `json:"totalTrades"`
	BasePrice    int64  `json:"basePrice"`
	MainCategory int    `json:"mainCategory"`
	SubCategory  int    `json:"subCategory"`
}

const apiBase = "https://api.arsha.io/v2/eu"

var httpClient = &http.Client{Timeout: 30 * time.Second}

func request(method, url string, body []byte, out any) error {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(data))
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w. %s", err, string(data))
	}
	return nil
}

func GetWorldMarketList(mainCategory int, subCategories ...int) ([]MarketItem, error) {
	fetch := func(url string) ([]MarketItem, error) {
		var items []MarketItem
		if err := request(http.MethodGet, url, nil, &items); err != nil {
			return nil, fmt.Errorf("world market list (%s): %w", url, err)
		}
		return items, nil
	}

	base := fmt.Sprintf("%s/GetWorldMarketList?lang=en&mainCategory=%d", apiBase, mainCategory)
	if len(subCategories) == 0 {
		return fetch(base)
	}

	var all []MarketItem
	for _, sub := range subCategories {
		items, err := fetch(fmt.Sprintf("%s&subCategory=%d", base, sub))
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
	}
	return all, nil
}

func GetBiddingInfo(items []MarketItem) ([]BiddingInfo, error) {
	var result []BiddingInfo
	for start := 0; start < len(items); start += 20 {
		end := min(start+20, len(items))
		ids := make([]PostId, 0, end-start)
		for _, item := range items[start:end] {
			ids = append(ids, PostId{Id: item.Id})
		}
		body, err := json.Marshal(ids)
		if err != nil {
			return nil, fmt.Errorf("marshal post body: %w", err)
		}
		var out []BiddingInfo
		if err := request(http.MethodPost, apiBase+"/GetBiddingInfoList?lang=en", body, &out); err != nil {
			return nil, fmt.Errorf("bidding info list: %w", err)
		}
		result = append(result, out...)
	}
	return result, nil
}
