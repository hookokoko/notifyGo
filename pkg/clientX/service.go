package clientX

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

var ServicesMap sync.Map

func NewServices(idc string) error {
	// 读conf目录下的所有文件
	tomlFilesPath, err := getTomlFile("conf")
	if err != nil {
		return err
	}

	// 针对每一个文件build service
	for _, path := range tomlFilesPath {
		config, err := loadService(path)
		if err != nil {
			log.Printf("加载%s时出现了错误%+v\n", path, err)
			continue
		}
		if err = buildService(idc, config); err != nil {
			log.Printf("创建%s-%s时出现了错误%+v\n", path, idc, err)
			continue
		}
	}

	return nil
}

func (s *Service) PickTarget() *Addr {
	addr, err := s.Bala.Next()
	if err != nil {
		log.Printf("pick addr fail, %s-%s", s.Name, s.Strategy)
		return nil
	}
	return addr
}

func (s *Service) GetTargets() ([]*Addr, error) {
	// 目前只接入手动指定host
	return s.Resource.Manual[s.IDC], nil
}

func getTomlFile(dirPath string) ([]string, error) {
	pathArr := make([]string, 0, 16)
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".toml") {
			pathArr = append(pathArr, path)
		}

		return nil
	})
	return pathArr, err
}

func loadService(path string) (*ServiceConfig, error) {
	var conf ServiceConfig
	_, err := toml.DecodeFile(path, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func buildService(idc string, config *ServiceConfig) error {
	var service *Service

	service = &Service{
		IDC:           idc,
		ServiceConfig: config,
	}

	err := buildBalancer(service)
	if err != nil {
		return err
	}

	ServicesMap.Store(config.Name, service)
	return nil
}

func buildBalancer(srv *Service) error {
	addrs, err := srv.GetTargets()
	if err != nil {
		return err
	}
	bala := NewBalanceBuilder[*Addr](srv.Name, addrs).Build(srv.Strategy)
	srv.Bala = bala
	return nil
}

func (a *Addr) GetReqDomain(isHttp bool) string {
	proto := "https"
	if isHttp {
		proto = "http"
	}
	if a.Port == "0" {
		return fmt.Sprintf("%s://%s", proto, a.Host)
	}
	return fmt.Sprintf("%s://%s:%s", proto, a.Host, a.Port)
}
