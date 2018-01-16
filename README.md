# qiita-team-feed

[![License](https://img.shields.io/github/license/masutaka/qiita-team-feed.svg?style=flat-square)][license]

[license]: https://github.com/masutaka/qiita-team-feed/blob/master/LICENSE.txt

任意の Qiita:Team のフィード（Atom 形式）を配送する。

## Development

WIP

1. 環境変数を設定する

    ```
    QIITA_ACCESS_TOKEN=<app.json を参考にする>
    QIITA_TEAM_NAME=<app.json を参考にする>
    REDIS_URL=redis://localhost:16379
    PORT=18080
    ```

1. Redis を起動する

    ```
    $ docker-compose up -d
    ```

1. user と token を Redis に保存する

    ```
    $ docker-compose exec redis redis-cli
    127.0.0.1:6379> set user:taro hogehoge
    OK
    127.0.0.1:6379> keys *
    1) "user:taro"
    ```

1. ビルド環境のセットアップ

    ```
    $ dep ensure
    ```

1. フィードを Redis に保存する

    ```
    $ go run main.go cli
    ```

1. HTTP サーバを起動する

    ```
    $ go run main.go
    ```

1. http://localhost:18080/feed?user=taro&token=hogehoge でフィードが取得できる

## Deploy to Heroku

WIP

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)

## Todo

* [x] フィードのアイテム数を環境変数でカスタマイズ可能にする
* [ ] Makefile を作る？
* [ ] リファクタリング
* [ ] テストを書く
* [ ] 正しい http status を返す
* [ ] Qiita ログインを実装して、当該 Qiita:Team メンバーは誰でもフィードを購読できるようにする
* [ ] 退職者からのアクセスは自動的に不可にする

## 実装方針

* 複数 Qiita:Team に対応する予定はない
* 本文は配送しない。URL 漏洩の影響を最小限に抑えるため
* 設置した人の権限によっては出力されるフィードには非公開記事も出力される？未確認
