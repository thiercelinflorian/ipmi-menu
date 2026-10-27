package ipmi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"
)

const NetbootISO = "http://boot.netboot.xyz/ipxe/netboot.xyz.iso"

var redfishBase = func(t Target) string { return "https://" + net.JoinHostPort(t.Host, "443") }

var mediaDelay = 2 * time.Second

var redfishClient = &http.Client{
	Timeout: Normal,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext:     (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
	},
}

func (t Target) redfish(method, path string, body, out any) error {
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, redfishBase(t)+path, r)
	if err != nil {
		return err
	}
	user := t.User
	if user == "" {
		user = DefaultUser
	}
	req.SetBasicAuth(user, t.Password)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := redfishClient.Do(req)
	if err != nil {
		return fmt.Errorf("Redfish injoignable sur %s (HTTPS 443)\n→ BMC trop ancien ou Redfish désactivé ?", t.Host)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return errors.New("Redfish : identifiants refusés par le BMC")
	case resp.StatusCode >= 300:
		return fmt.Errorf("Redfish %s %s : %s", method, path, redfishError(resp.Status, data))
	case out != nil:
		return json.Unmarshal(data, out)
	}
	return nil
}

func redfishError(status string, data []byte) string {
	var e struct {
		Error struct {
			Message  string `json:"message"`
			Extended []struct {
				Message string `json:"Message"`
			} `json:"@Message.ExtendedInfo"`
		} `json:"error"`
	}
	json.Unmarshal(data, &e)
	if len(e.Error.Extended) > 0 && e.Error.Extended[0].Message != "" {
		return e.Error.Extended[0].Message
	}
	if e.Error.Message != "" {
		return e.Error.Message
	}
	return status
}

type link struct {
	ID string `json:"@odata.id"`
}

type virtualMedia struct {
	ID         string   `json:"@odata.id"`
	MediaTypes []string `json:"MediaTypes"`
	Inserted   bool     `json:"Inserted"`
	Actions    struct {
		Insert link `json:"#VirtualMedia.InsertMedia"`
		Eject  link `json:"#VirtualMedia.EjectMedia"`
	} `json:"Actions"`
}

func (l *link) UnmarshalJSON(b []byte) error {
	var v struct {
		ID     string `json:"@odata.id"`
		Target string `json:"target"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	l.ID = v.ID
	if v.Target != "" {
		l.ID = v.Target
	}
	return nil
}

func (t Target) virtualCD() (virtualMedia, error) {
	var parents []string
	for _, base := range []string{"/redfish/v1/Managers", "/redfish/v1/Systems"} {
		var c struct{ Members []link }
		if err := t.redfish("GET", base, nil, &c); err != nil {
			if base == "/redfish/v1/Managers" {
				return virtualMedia{}, err
			}
			continue
		}
		for _, m := range c.Members {
			parents = append(parents, m.ID+"/VirtualMedia")
		}
	}
	for _, p := range parents {
		var c struct{ Members []link }
		if t.redfish("GET", p, nil, &c) != nil {
			continue
		}
		for _, m := range c.Members {
			var v virtualMedia
			if t.redfish("GET", m.ID, nil, &v) != nil {
				continue
			}
			if slices.Contains(v.MediaTypes, "CD") || slices.Contains(v.MediaTypes, "DVD") {
				return v, nil
			}
		}
	}
	return virtualMedia{}, errors.New("ce BMC n'expose pas de CD virtuel en Redfish")
}

func (t Target) InsertISO(url string) error {
	v, err := t.virtualCD()
	if err != nil {
		return err
	}
	if v.Inserted {
		t.eject(v)
	}
	if a := v.Actions.Insert.ID; a != "" {
		err := t.redfish("POST", a, map[string]any{"Image": url, "Inserted": true, "WriteProtected": true}, nil)
		if err != nil && strings.Contains(err.Error(), "Redfish POST") {
			err = t.redfish("POST", a, map[string]any{"Image": url}, nil)
		}
		return err
	}
	return t.redfish("PATCH", v.ID, map[string]any{"Image": url, "Inserted": true}, nil)
}

func (t Target) EjectISO() error {
	v, err := t.virtualCD()
	if err != nil || !v.Inserted {
		return err
	}
	return t.eject(v)
}

func (t Target) eject(v virtualMedia) error {
	if a := v.Actions.Eject.ID; a != "" {
		return t.redfish("POST", a, map[string]any{}, nil)
	}
	return t.redfish("PATCH", v.ID, map[string]any{"Image": nil, "Inserted": false}, nil)
}

func (t Target) Netboot(iso string, legacy, reboot bool) (dropped bool, err error) {
	if iso == "" {
		iso = NetbootISO
	}
	if err := t.InsertISO(iso); err != nil {
		return false, err
	}
	time.Sleep(mediaDelay)
	if dropped, err = t.SetBoot("cdrom", BootOptions{Legacy: legacy}); err != nil {
		return dropped, err
	}
	if reboot {
		err = t.Reboot()
	}
	return dropped, err
}
