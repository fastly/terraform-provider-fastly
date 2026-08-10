# Fastly Terraform Provider

<!-- hy-mt2-i18n:start -->
[English](./README.md) | [中文](./README_zh-CN.md) | **日本語** | [Español](./README_es.md)
<!-- hy-mt2-i18n:end -->


- ウェブサイト: https://www.terraform.io  
- ドキュメント: https://registry.terraform.io/providers/fastly/fastly/latest/docs  
- 問題報告: https://github.com/fastly/terraform-provider-fastly/blob/main/ISSUES.md  
- メーリングリスト: http://groups.google.com/group/terraform-tool  
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)

## 要件

- [Terraform](https://www.terraform.io/downloads.html) 0.12.x 以降  
- [Go](https://golang.org/doc/install) 1.25（プロバイダープラグインのビルド用）

> 注意：Terraform 0.11.x およびそれ以前をサポートしていた最終バージョンの Fastly プロバイダは [v0.26.0](https://github.com/fastly/terraform-provider-fastly/releases/tag/v0.26.0) です。

## バージョン管理とリリーススケジュール

このモジュールのメンテナーたちは、[セマンティックバージョニング
(SemVer)](https://semver.org/)を厳守しようと努めています。つまり、互換性を破壊する変更
（機能の削除、または既存機能に対する非互換な変更）は、最初のバージョン要素
（`major`）がインクリメントされたバージョンでリリースされます。新機能の追加は
2番目のバージョン要素（`minor`）をインクリメントし、互換性に影響を与えない
バグ修正は3番目のバージョン要素（`patch`）をインクリメントします。

毎月3番目の水曜日に、リリースする準備が整ったすべての破壊的変更、機能追加、バグ修正が含まれたバージョンが公開されます。もしその水曜日がアメリカの祝日と重なった場合、リリースは次の営業日まで延期されます。

これらのメジャーリリースの間に重要または緊急なバグ修正がリリース可能になった場合、それらの修正を適用するために必要に応じてパッチリリースが行われます。

## プロバイダのビルド

リポジトリを以下の場所にクローンします：`$GOPATH/src/github.com/fastly/terraform-provider-fastly`

```sh
$ mkdir -p $GOPATH/src/github.com/fastly; cd $GOPATH/src/github.com/fastly
$ git clone git@github.com:fastly/terraform-provider-fastly
```

プロバイダーのディレクトリに移動し、プロバイダーをビルドします。

```sh
$ cd $GOPATH/src/github.com/fastly/terraform-provider-fastly
$ make build
```

## プロバイダの開発

プロバイダの開発を行う場合、まずご使用のマシンに[Go](http://www.golang.org)がインストールされている必要があります（バージョン1.25以上が*必須*です）。

プロバイダをコンパイルするには、`make build` を実行してください。これによりプロバイダがビルドされ、そのバイナリがローカルの `bin` ディレクトリに配置されます。

```sh
$ make build
...
```

新たにビルドされたバイナリと共に、`developer_overrides.tfrc` という名前のファイルが作成されます。`make build` ターゲットは、Terraform がローカルでビルドされたプロバイダーバイナリを使用できるようにするための `TF_CLI_CONFIG_FILE` 環境変数の設定方法に関する詳細を返します。

- HashiCorp - [Provider開発者向けの開発時オーバーライド](https://www.terraform.io/docs/cli/config/config-file.html#development-overrides-for-provider-developers)。

> **注記**: プロバイダーに加えたコード変更の影響が確認できない場合、Terraform CLIがどのプロバイダーバイナリを使用すべきか判断できず混乱している可能性があります。`./bin/`ディレクトリ内を確認し、異なるコミットハッシュを持つ複数のプロバイダー（例: `terraform-provider-fastly_v2.2.0-5-gfdc37cee`）が存在する場合は、`make build`を実行する前にそれらを削除してください。これにより、Terraform CLIが正しいバイナリを選択できるようになります。

### プロバイダのデバッグ

`dev_overrides` を使用する前述の方法は、実際の Terraform コードを使ってローカルで変更内容をテストするなど、ほとんどの開発用途において十分でしょう。
ただし、特定の問題を解決するために [delve](https://github.com/go-delve/delve) のようなデバッガを接続する必要がある場合は、デバッグモードでプロバイダを実行する方が役立つこともあります。

Terraformの通常の動作方式としては、サブプロセスでプロバイダを起動し、ローカルソケット経由のGRPCを使ってそれに接続します。
（詳細については、[hashicorp/go-plugin](https://github.com/hashicorp/go-plugin#architecture)を参照してください。）
デバッグ時には、この方式を迂回してプロバイダを別のプロセスで実行し、Terraformにそのプロセスとの通信方法を指示することが可能です。
これによる利点は、デバッガを通常の実行ファイルに添付するのと同じ方法で、デバッガを添付した状態でプロバイダを起動できる点です。

使用するデバッガーによって実行方法はいくつかありますが、ここでは[delve](https://github.com/go-delve/delve)を使用します。なぜなら、他のデバッガーでも手順はほぼ同じだからです。設定が必要なのは2点で、1つ目はプロバイダーを最適化なしでコンパイルすること、もう1つは実行ファイルを`--debug`フラグを付けて起動することです。最適化なしでコンパイルすることで、デバッガーがバイナリ内にある必要なすべてのシンボルにアクセスできるようになり、`--debug`フラグはTerraformプラグインSDKに対し、別途プロセスとして実行されること、および起動後に接続手順が表示されることを伝えます。

[delve](https://github.com/go-delve/delve) を使用すれば、これは単一のコマンドで実行できます：

```sh
$ dlv debug. -- --debug
コマンド一覧を表示するには 'help' を入力してください。
(dlv) continue
{"@level":"debug","@message":"plugin address","@timestamp":"2021-03-26T12:10:13.320981Z","address":"/var/folders/qm/swg2hf4h5t8sdht8yhds4dg6m0000gn/T/plugin865249851","network":"unix"}
プロバイダーが起動しました。Terraform を接続するには TF_REATTACH_PROVIDERS 環境変数を設定してください：

        TF_REATTACH_PROVIDERS='{"fastly/fastly":{"Protocol":"grpc","Pid":54132,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/qm/swg2hf4h5t8sdht8yhds4dg6m0000gn/T/plugin865249851"}}}'
```

これは2つの別々なステップで行うことも可能です。`-gcflags`オプションにより最適化（`-N`）とインライン処理（`-l`）が無効になります。

```sh
$ go build -gcflags="all=-N -l" -o terraform-provider-fastly_debug
$ dlv exec terraform-provider-fastly_debug -- --debug
```

メッセージに従って別のシェルに移動し、`TF_REATTACH_PROVIDERS`という環境変数をエクスポートしてください。その後、通常通りTerraformを使用すると、デバッガー内で自動的にこのプロバイダが利用されます。

```sh
$ export TF_REATTACH_PROVIDERS='{"fastly/fastly":{"Protocol":"grpc","Pid":54132,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/qm/swg2hf4h5t8sdht8yhds4dg6m0000gn/T/plugin865249851"}}}'
$ terraform plan
```

その後、通常通りデバッガーを使ってブレークポイントを設定し、プロバイダーの実行経過をトレースできるようになります。

デバッグモードの設定方法は、Terraform 0.13.x の使用を前提としています。Terraform 0.12.x を使用している場合は、`TF_REATTACH_PROVIDERS` に割り当てられる値を手動で変更し、キー `"fastly/fastly"` を `"registry.terraform.io/-/fastly"` にする必要があります。詳細については、HashiCorp の [「デバッグ可能なプロバイダーバイナリのサポート」](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#support-for-debuggable-provider-binaries) を参照してください。

### まとめ

```
（最初のシェル）dlv debug. --headless -- --debug
（2番目のシェル）dlv connect <最初のシェルからの出力>
               continue
               <Ctrl-c>
               fastly/block_fastly_service_package.go:123 で停止
（3番目のシェル）export TF_REATTACH_PROVIDERS="..."
               terraform apply
（2番目のシェル）continue（ブレークポイント検証を続行）
               <Ctrl-c>（その後、3番目のシェルから別のterraformコマンドを実行）
```

## テスト

プロバイダをテストするには、単純に `make test` を実行すればよいです。

```sh
$ make test
```

全ての受容テストを実行するには、`make testacc` を実行してください。

*注:* アクセプタンステストでは実際のリソースが作成され、実行には多額のコストがかかることがあります。全てのアクセプタンステストを実行するには数時間かかると考えておくべきです。

```sh
$ make testacc
```

個別のアクセプタンステストを実行するには、正規表現と一緒に‘-run’フラグを使用できます。
以下の例では、‘TestAccFastlyServiceVCL_basic’という単一のテストにマッチする正規表現が使われています。

```sh
$ make testacc TESTARGS='-run=TestAccFastlyServiceVCL_basic'
```

以下の例では、正規表現を使用して一連の基本的なアクセプタンステストを実行しています。

```sh
$ make testacc TESTARGS='-run=TestAccFastlyServiceVCL.*_basic'
```

より詳細なデバッグ情報を持ってテストを実行するには、`make` コマンドの前に `TF_LOG` を付けてください（詳細は [terraform のドキュメント](https://www.terraform.io/docs/internals/debugging.html) を参照してください）。

```sh
$ TF_LOG=trace make testacc
```

デフォルトでは、テストは並列度4で実行されます。
ネットワーク関連の問題により一部のテストが失敗した場合はこの値を下げ、テストの実行時間を短縮するために可能であれば上げることもできます。
これを設定するには、次の例のように `make` コマンドの前に `TEST_PARALLELISM` を付けてください。

```sh
$ TEST_PARALLELISM=8 make testacc
```

使用しているFastlyアカウントによっては、Platform TLSなどの機能が有効になっていない場合があります。
これにより、全テストセットを実行する際に一部のテストが失敗し、`403 Unauthorised`エラーが発生する可能性があります。
失敗したテストが限定提供されている機能や特定の顧客のみが利用できる機能を使用しているかどうかを確認するには、[Fastly APIドキュメント](https://developer.fastly.com/reference/api/)を参照してください。
そのような場合は、上記の`TESTARGS`正規表現を使用するか、除外すべきテストの先頭に一時的に`t.SkipNow()`を追加してください。

## ドキュメントのビルド

[ドキュメントガイド](./DOCUMENTATION.md)を参照してください。

## 貢献方法

[CONTRIBUTING.md](./CONTRIBUTING.md) を参照してください。
