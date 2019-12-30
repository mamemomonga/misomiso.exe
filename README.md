# misomiso.exe

みそみそ〜、とマストドンでつぶやく専用コマンドラインアプリケーションです。

# 利用方法

* [etc/config-example.yaml](etc/config-example.yaml)を参考にconfig.yamlを作成してください。

実行例

	./bin/misomiso -config ./etc/config.yaml -target 'ゆゆ式' -regexp '(ゆゆ(式|しき)|yysk|yuyush?iki)'

# コマンドラインオプション

* -config  Configファイルを指定します
* -target  キーワードを指定します
* -regexp  正規表現を指定します

# 開発

## ビルド環境

事前に必要なもの

* make
* go

### 準備

* GOPATHが正しく設定されている必要があります。

コマンド

	$ make


