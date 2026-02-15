package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagernet/sing-box/common/geosite"
	"github.com/sagernet/sing-box/common/srs"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

// ConvertGeoIP reads geoip.dat, filters by the given categories,
// and writes .srs files for each category into outDir.
func ConvertGeoIP(datPath string, categories []string, outDir string) error {
	data, err := os.ReadFile(datPath)
	if err != nil {
		return fmt.Errorf("failed to read geoip.dat: %w", err)
	}

	var geoipList routercommon.GeoIPList
	if err := proto.Unmarshal(data, &geoipList); err != nil {
		return fmt.Errorf("failed to parse geoip.dat: %w", err)
	}

	// Build a set of requested categories (lowercase)
	categorySet := make(map[string]bool, len(categories))
	for _, c := range categories {
		categorySet[strings.ToLower(c)] = true
	}

	found := make(map[string]bool)

	for _, entry := range geoipList.Entry {
		code := strings.ToLower(entry.GetCountryCode())
		if code == "" || !categorySet[code] {
			continue
		}
		found[code] = true

		var headlessRule option.DefaultHeadlessRule
		headlessRule.IPCIDR = make([]string, 0, len(entry.Cidr))

		for _, cidr := range entry.Cidr {
			headlessRule.IPCIDR = append(headlessRule.IPCIDR,
				net.IP(cidr.GetIp()).String()+"/"+fmt.Sprint(cidr.GetPrefix()))
		}

		headlessRule.IPCIDR = common.Uniq(headlessRule.IPCIDR)

		plainRuleSet := option.PlainRuleSet{
			Rules: []option.HeadlessRule{
				{
					Type:           C.RuleTypeDefault,
					DefaultOptions: headlessRule,
				},
			},
		}

		srsPath := filepath.Join(outDir, "geoip-"+code+".srs")
		if err := writeSRS(srsPath, plainRuleSet); err != nil {
			return fmt.Errorf("failed to write %s: %w", srsPath, err)
		}
		fmt.Printf("  ✓ geoip-%s.srs\n", code)
	}

	// Warn about categories not found
	for _, c := range categories {
		if !found[strings.ToLower(c)] {
			fmt.Printf("  ⚠ geoip category %q not found in geoip.dat\n", c)
		}
	}

	return nil
}

// ConvertGeoSite reads geosite.dat, filters by the given categories,
// and writes .srs files for each category into outDir.
func ConvertGeoSite(datPath string, categories []string, outDir string) error {
	data, err := os.ReadFile(datPath)
	if err != nil {
		return fmt.Errorf("failed to read geosite.dat: %w", err)
	}

	var geositeList routercommon.GeoSiteList
	if err := proto.Unmarshal(data, &geositeList); err != nil {
		return fmt.Errorf("failed to parse geosite.dat: %w", err)
	}

	// Build a set of requested categories (lowercase)
	categorySet := make(map[string]bool, len(categories))
	for _, c := range categories {
		categorySet[strings.ToLower(c)] = true
	}

	// Parse all entries into domain map (including attribute-based sub-categories)
	domainMap := make(map[string][]geosite.Item)
	for _, entry := range geositeList.Entry {
		code := strings.ToLower(entry.CountryCode)
		domains := make([]geosite.Item, 0, len(entry.Domain)*2)
		attributes := make(map[string][]*routercommon.Domain)

		for _, domain := range entry.Domain {
			if len(domain.Attribute) > 0 {
				for _, attr := range domain.Attribute {
					attributes[attr.Key] = append(attributes[attr.Key], domain)
				}
			}

			switch domain.Type {
			case routercommon.Domain_Plain:
				domains = append(domains, geosite.Item{
					Type:  geosite.RuleTypeDomainKeyword,
					Value: domain.Value,
				})
			case routercommon.Domain_Regex:
				domains = append(domains, geosite.Item{
					Type:  geosite.RuleTypeDomainRegex,
					Value: domain.Value,
				})
			case routercommon.Domain_RootDomain:
				if strings.Contains(domain.Value, ".") {
					domains = append(domains, geosite.Item{
						Type:  geosite.RuleTypeDomain,
						Value: domain.Value,
					})
				}
				domains = append(domains, geosite.Item{
					Type:  geosite.RuleTypeDomainSuffix,
					Value: "." + domain.Value,
				})
			case routercommon.Domain_Full:
				domains = append(domains, geosite.Item{
					Type:  geosite.RuleTypeDomain,
					Value: domain.Value,
				})
			}
		}
		domainMap[code] = common.Uniq(domains)

		for attr, attrEntries := range attributes {
			attrDomains := make([]geosite.Item, 0, len(attrEntries)*2)
			for _, domain := range attrEntries {
				switch domain.Type {
				case routercommon.Domain_Plain:
					attrDomains = append(attrDomains, geosite.Item{
						Type:  geosite.RuleTypeDomainKeyword,
						Value: domain.Value,
					})
				case routercommon.Domain_Regex:
					attrDomains = append(attrDomains, geosite.Item{
						Type:  geosite.RuleTypeDomainRegex,
						Value: domain.Value,
					})
				case routercommon.Domain_RootDomain:
					if strings.Contains(domain.Value, ".") {
						attrDomains = append(attrDomains, geosite.Item{
							Type:  geosite.RuleTypeDomain,
							Value: domain.Value,
						})
					}
					attrDomains = append(attrDomains, geosite.Item{
						Type:  geosite.RuleTypeDomainSuffix,
						Value: "." + domain.Value,
					})
				case routercommon.Domain_Full:
					attrDomains = append(attrDomains, geosite.Item{
						Type:  geosite.RuleTypeDomain,
						Value: domain.Value,
					})
				}
			}
			domainMap[code+"@"+attr] = common.Uniq(attrDomains)
		}
	}

	// Convert only requested categories
	found := make(map[string]bool)
	for _, cat := range categories {
		code := strings.ToLower(cat)
		domains, ok := domainMap[code]
		if !ok {
			fmt.Printf("  ⚠ geosite category %q not found in geosite.dat\n", cat)
			continue
		}
		found[code] = true

		var headlessRule option.DefaultHeadlessRule
		compiled := geosite.Compile(domains)
		headlessRule.Domain = compiled.Domain
		headlessRule.DomainSuffix = compiled.DomainSuffix
		headlessRule.DomainKeyword = compiled.DomainKeyword
		headlessRule.DomainRegex = compiled.DomainRegex

		plainRuleSet := option.PlainRuleSet{
			Rules: []option.HeadlessRule{
				{
					Type:           C.RuleTypeDefault,
					DefaultOptions: headlessRule,
				},
			},
		}

		srsPath := filepath.Join(outDir, "geosite-"+code+".srs")
		if err := writeSRS(srsPath, plainRuleSet); err != nil {
			return fmt.Errorf("failed to write %s: %w", srsPath, err)
		}
		fmt.Printf("  ✓ geosite-%s.srs\n", code)
	}

	return nil
}

func writeSRS(path string, ruleSet option.PlainRuleSet) error {
	absPath, _ := filepath.Abs(path)
	outFile, err := os.Create(absPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return srs.Write(outFile, ruleSet, C.RuleSetVersion2)
}
