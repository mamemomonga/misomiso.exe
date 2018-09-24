# misomiso.exe

みそみそ〜、とマストドンでつぶやく専用コマンドラインアプリケーションです。

# 便利な使い方(Windows)

1. [こちらから](https://github.com/mamemomonga/misomiso.exe/releases) misomiso.exe をダウンロードします。

2. misomiso.exe と同じフォルダに
以下のような内容を自分のマストドンログイン情報に書き換えます

config.yaml という名前で保存します。

	mastodon:
	   domain: mstdn.jp
	   email: example@example.com
	   password: password

ファイルはYAML形式です。インデントは必須でタブ文字は使えません。

3. misomiso.exe をダブルクリックすると、「みそみそ〜」と投稿されます。

4. そのあとLTLとHTLで発見したすべての「みそみそ」を含むトゥートをブーストとファボします。(みそチェイサー機能)

5. 60秒間それらのトゥートを発見できなかったら終了します

## コマンドラインオプション

	--help ヘルプ表示
	-r     検索用正規表現設定
	-t     開始宣言文言設定


# ビルド

## 必要なもの

* Docker
* Bash

macOS High Sierra にて動作確認を行っています。

## ビルド方法

build.sh のまん中あたりにあるところで、ビルドしたいターゲットのコメントを外してください。

	$ ./build.sh

