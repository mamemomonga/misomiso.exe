# misomiso.exe

みそみそ〜、とマストドンでつぶやく専用コマンドラインアプリケーションです。

# 利用方法

* [リリースページ](https://github.com/mamemomonga/misomiso.exe/releases) からバイナリをダウンロードします。
* Linux, macOSの場合は実行権限をつけてください。
* Windowsの場合はセキュリティー警告が表示される場合があります。
* [etc/config-example.yaml](etc/config-example.yaml)を参考にconfig.yamlを作成してください。

実行例

	./misomiso-darwin-amd64 -config ./etc/config.yaml -target 'ゆゆ式' -regexp '(ゆゆ(式|しき)|yysk|yuyush?iki)'

# コマンドラインオプション

* -config  Configファイルを指定します
* -target  キーワードを指定します
* -regexp  正規表現を指定します

# 開発

## ビルド環境

事前に必要なもの

* make
* go

導入されてない場合導入されるもの
		
* dep

### 準備

* GOPATHが正しく設定されている必要があります。
* make deps を実行すると、dep eusure が実行されます。depが導入されていない場合は導入されます。

コマンド

	$ git clone https://github.com/mamemomonga/misomiso.exe $GOPATH/src/github.com/mamemomonga/misomiso.exe
	$ cd $GOPATH/src/github.com/mamemomonga/misomiso.exe
	$ make deps
	$ make run
	$ make

### リリース向けビルド

* 公開用のバイナリが生成されます。
* Dockerが必要です。

コマンド

	$ make release

