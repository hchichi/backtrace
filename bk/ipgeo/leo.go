package ipgeo

import (
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/hchichi/backtrace/bk/wshandle"
	"github.com/tidwall/gjson"
)

type IPGeoData struct {
	Asnumber  string
	Country   string
	CountryEn string
	Prov      string
	ProvEn    string
	City      string
	CityEn    string
	District  string
	Owner     string
	Lat       float64
	Lng       float64
	Isp       string
	Whois     string
	Prefix    string
	Router    map[string][]string
}

type IPPool struct {
	pool    map[string]chan IPGeoData
	poolMux sync.Mutex
	// 添加清理计时器
	cleanupTimer *time.Timer
}

var IPPools = IPPool{
	pool: make(map[string]chan IPGeoData),
}

func init() {
	// 启动清理协程
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			IPPools.cleanup()
		}
	}()
}

func (p *IPPool) cleanup() {
	p.poolMux.Lock()
	defer p.poolMux.Unlock()

	// 清理所有通道
	for ip, ch := range p.pool {
		select {
		case <-ch:
			// 通道有数据，清空它
		default:
			// 通道为空或已关闭
		}
		close(ch)
		delete(p.pool, ip)
	}
}

func (p *IPPool) getChannel(ip string) chan IPGeoData {
	p.poolMux.Lock()
	defer p.poolMux.Unlock()

	if p.pool[ip] == nil {
		p.pool[ip] = make(chan IPGeoData, 1) // 使用带缓冲的通道
	}
	return p.pool[ip]
}

func (p *IPPool) removeChannel(ip string) {
	p.poolMux.Lock()
	defer p.poolMux.Unlock()

	if ch, ok := p.pool[ip]; ok {
		close(ch)
		delete(p.pool, ip)
	}
}

func sendIPRequest(ip string) error {
	wsConn, err := wshandle.GetWsConn()
	if err != nil {
		return err
	}
	wsConn.MsgSendCh <- ip
	return nil
}

func receiveParse() error {
	wsConn, err := wshandle.GetWsConn()
	if err != nil {
		return err
	}

	wsConn.ConnMux.Lock()
	defer wsConn.ConnMux.Unlock()

	select {
	case data := <-wsConn.MsgReceiveCh:
		res := gjson.Parse(data)
		
		var domain = res.Get("domain").String()
		if domain == "" {
			domain = res.Get("owner").String()
		}

		m := make(map[string][]string)
		err := json.Unmarshal([]byte(res.Get("router").String()), &m)
		if err != nil {
			// 某些 IP 没有路由信息，这是正常的
		}

		lat, _ := strconv.ParseFloat(res.Get("lat").String(), 32)
		lng, _ := strconv.ParseFloat(res.Get("lng").String(), 32)

		IPPools.pool[res.Get("ip").String()] <- IPGeoData{
			Asnumber:  res.Get("asnumber").String(),
			Country:   res.Get("country").String(),
			CountryEn: res.Get("country_en").String(),
			Prov:      res.Get("prov").String(),
			ProvEn:    res.Get("prov_en").String(),
			City:      res.Get("city").String(),
			CityEn:    res.Get("city_en").String(),
			District:  res.Get("district").String(),
			Owner:     domain,
			Lat:       lat,
			Lng:       lng,
			Isp:       res.Get("isp").String(),
			Whois:     res.Get("whois").String(),
			Prefix:    res.Get("prefix").String(),
			Router:    m,
		}
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("receive timeout")
	}
}

func LeoIP(ip string, timeout time.Duration) (*IPGeoData, error) {
	if timeout < 5*time.Second {
		timeout = 5 * time.Second
	}

	// 获取或创建通道
	ch := IPPools.getChannel(ip)
	defer IPPools.removeChannel(ip) // 请求完成后清理通道

	err := sendIPRequest(ip)
	if err != nil {
		return nil, err
	}

	errCh := make(chan error, 1) // 使用带缓冲的通道
	go func() {
		errCh <- receiveParse()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
		select {
		case res := <-ch:
			return &res, nil
		case <-time.After(timeout):
			return nil, errors.New("timeout waiting for response")
		}
	case <-time.After(timeout):
		return nil, errors.New("timeout waiting for parse")
	}
}
