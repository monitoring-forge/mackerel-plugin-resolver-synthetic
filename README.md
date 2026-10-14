# mackerel-plugin-resolver-synthetic

DNSサーバーに名前解決を問い合わせ、応答状況と応答時間を監視するMackerelプラグインです。

## インストール

`mkr`を使ってGitHubからインストールします。

```sh
mkr plugin install github.com/monitoring-forge/mackerel-plugin-resolver-synthetic
```

## 使い方

コマンドを直接実行すると、標準出力にMackerel形式のメトリックを出力します。

```sh
mackerel-plugin-resolver-synthetic --hostname 8.8.8.8 --hostname 8.8.4.4 --question example.com
```

Mackerel Agentから定期実行する場合は、エージェントの設定ファイルにプラグインを登録します。必要に応じてコマンドにオプションを追加してください。

```toml
[plugin.metrics.resolver-synthetic]
command = "mackerel-plugin-resolver-synthetic --hostname 8.8.8.8 --hostname 8.8.4.4 --question example.com"
```

## オプション

| オプション | 説明 | デフォルト |
| --- | --- | --- |
| `-H`, `--hostname` | 問い合わせ先のDNSサーバー。複数指定できます。 | `127.0.0.1` |
| `-Q`, `--question` | Aレコードを問い合わせるホスト名。末尾の `.` は省略できます。 | `example.com.` |
| `-E`, `--expect` | DNS応答のAレコードに含まれることを確認する文字列。 | なし |
| `-R`, `--expect-regexp` | DNS応答のAレコードに一致することを確認する正規表現。`--expect` とは同時に指定できません。 | なし |
| `--timeout` | DNS問い合わせのタイムアウト。問い合わせ先が複数の場合、問い合わせ順に応じて各問い合わせの値が調整されます。 | `5s` |
| `--attempts` | DNSサーバーの一覧を最大何周するか。いずれかの問い合わせが成功すると終了します。 | `2` |
| `--deadline` | 一連の問い合わせ全体に対する制限時間。 | `20s` |
| `--prefix` | メトリックキーのprefix。メトリック名には末尾に `-synthetic` が付加されます。 | `resolver` |
| `-v`, `--version` | バージョンを表示します。 | - |
| `-h`, `--help` | ヘルプを表示します。 | - |

## 出力メトリック

出力は `メトリック名 値 Unix時刻` の形式です。prefixのデフォルトは `resolver-synthetic` です。

| メトリック | 単位 | 意味 |
| --- | --- | --- |
| `resolver-synthetic.service.available` | パーセント | DNS応答が成功し、指定した `--expect` または `--expect-regexp` の条件も満たした場合は `100`、それ以外は `0`。 |
| `resolver-synthetic.rtt.milliseconds` | ミリ秒 | 成功または処理終了までに実施したDNS問い合わせの所要時間の合計。失敗した問い合わせの時間も含みます。 |

問い合わせに失敗した場合、詳細は標準エラーにログ出力されます。

## ライセンス

[MIT License](LICENSE)
```
