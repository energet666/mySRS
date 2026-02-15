# srs-generator

Утилита для генерации SRS (sing-box rule set) файлов из `geoip.dat` и `geosite.dat`.

Загружает последние версии dat-файлов из релизов [runetfreedom/russia-v2ray-rules-dat](https://github.com/runetfreedom/russia-v2ray-rules-dat) и конвертирует указанные категории в формат `.srs`.

## Сборка

```bash
go build -o srs-generator .
```

## Использование

```bash
# С конфигом по умолчанию (config.yaml)
./srs-generator

# С указанием пути к конфигу
./srs-generator -config /path/to/config.yaml
```

## Конфигурация

Файл `config.yaml`:

```yaml
output_dir: "./output"

geosite:
  - ru-blocked
  - youtube
  - google
  - discord
  - category-ads-all

geoip:
  - ru-blocked
  - telegram
  - cloudflare
```

Полный список доступных категорий — в [README репозитория](https://github.com/runetfreedom/russia-v2ray-rules-dat).

Для удобного просмотра категорий в geosite и geoip можно использовать [GeoFiles Viewer](https://geofilesviewer.vpnvezdehod.com/).
