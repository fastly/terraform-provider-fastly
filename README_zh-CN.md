# Fastly Terraform 提供程序

<!-- hy-mt2-i18n:start -->
[English](./README.md) | **中文** | [日本語](./README_ja.md) | [Español](./README_es.md)
<!-- hy-mt2-i18n:end -->


- 网站：https://www.terraform.io  
- 文档：https://registry.terraform.io/providers/fastly/fastly/latest/docs  
- 问题报告：https://github.com/fastly/terraform-provider-fastly/blob/main/ISSUES.md  
- 邮件列表：http://groups.google.com/group/terraform-tool  
- [![Gitter 聊天](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)

## 要求条件

- [Terraform](https://www.terraform.io/downloads.html) 0.12.x 及更高版本
- [Go](https://golang.org/doc/install) 1.25（用于构建提供程序插件）

> 注意：最后一个支持 Terraform 0.11.x 及更低版本的 Fastly 提供程序版本是 [v0.26.0](https://github.com/fastly/terraform-provider-fastly/releases/tag/v0.26.0)。

## 版本控制与发布计划

该模块的维护者致力于遵循[语义化版本控制（SemVer）](https://semver.org/)规范。这意味着那些会造成兼容性问题的变更（如功能删除，或对现有功能进行的不兼容修改）将会通过版本号中第一个组件（“主版本号”）的递增来发布；新功能的添加则会使第二个组件（“次版本号”）递增；而不会影响兼容性的错误修复则会使第三个组件（“修订号”）递增。

在每个月的第三个星期三，会发布一个版本，其中包含所有已准备好发布的破坏性变更、功能新增以及错误修复。如果该星期三恰好是美国的节假日，发布将会推迟到下一个可行的工作日。

在两次主要版本发布之间，若有关键或紧急的漏洞修复需要发布，将会根据需求推出补丁版本以提供这些修复。

## 构建提供程序

将仓库克隆到：`$GOPATH/src/github.com/fastly/terraform-provider-fastly`

```sh
$ mkdir -p $GOPATH/src/github.com/fastly; cd $GOPATH/src/github.com/fastly
$ git clone git@github.com:fastly/terraform-provider-fastly
```

进入提供商目录并编译该提供商。

```sh
$ cd $GOPATH/src/github.com/fastly/terraform-provider-fastly
$ make build
```

## 开发该提供程序

如果您希望参与该提供程序的开发，首先需要在您的机器上安装 [Go](http://www.golang.org)（必须使用 1.25 及以上版本）。

要编译该提供程序，请运行 `make build`。该命令会构建提供程序，并将生成的二进制文件放入本地的 `bin` 目录中。

```sh
$ make build
...
```

在新建的二进制文件旁，还会生成一个名为 `developer_overrides.tfrc` 的文件。`make build` 目标会反馈用于设置 `TF_CLI_CONFIG_FILE` 环境变量的相关细节，该环境变量可让 Terraform 使用您本地编译生成的提供程序二进制文件。

- HashiCorp - [供提供商开发者使用的开发模式覆盖配置](https://www.terraform.io/docs/cli/config/config-file.html#development-overrides-for-provider-developers)。

> **注意**：如果您发现对提供程序所做的代码更改并未产生预期效果，那可能是 Terraform CLI 无法确定应使用哪个提供程序二进制文件。请检查 `./bin/` 目录中是否存在具有不同提交哈希值的多个提供程序（例如 `terraform-provider-fastly_v2.2.0-5-gfdc37cee`），并在运行 `make build` 之前先将它们删除。这样有助于 Terraform CLI 找到正确的二进制文件。

### 调试提供程序

对于大多数开发场景，包括使用实际的 Terraform 代码来测试本地所做的修改，之前通过 `dev_overrides` 实现的方法就已经足够了。不过，如果需要借助诸如 [delve](https://github.com/go-delve/delve) 这样的调试器来解决特定问题，那么以调试模式运行提供程序往往也会很有帮助。

Terraform 的常规工作方式是会在一个子进程中启动提供程序，并通过本地套接字使用 GRPC 与其建立连接。
（更多相关信息请参见 [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin#architecture) 。）
为了进行调试，可以绕过这一机制，在单独的进程中运行提供程序，然后再告知 Terraform 如何与该进程通信。
这样做的优势在于，可以像为普通可执行文件附加调试器那样，直接在启动提供程序时一并附加调试器。

根据所使用的调试器不同，实现这一目标的方法也有几种，但这里我们将以 [delve](https://github.com/go-delve/delve) 为例，因为其他调试器的操作流程应该大致相似。需要配置的两点分别是：以不带优化的方式编译提供程序，以及使用 `--debug` 参数运行该可执行文件。不进行优化编译能确保调试器能够访问二进制文件中的所有所需符号，而 `--debug` 参数则会让 Terraform 插件 SDK 明确知晓它将在独立的进程中运行，并在启动后显示连接指南。

使用 [delve](https://github.com/go-delve/delve) 时，只需一条命令即可完成此操作：

```sh
$ dlv debug. -- --debug
输入 ‘help’ 可查看命令列表。
(dlv) continue
{"@level":"debug","@message":"插件地址","@timestamp":"2021-03-26T12:10:13.320981Z","address":"/var/folders/qm/swg2hf4h5t8sdht8yhds4dg6m0000gn/T/plugin865249851","network":"unix"}
插件已启动，若要附加 Terraform，请设置 TF_REATTACH_PROVIDERS 环境变量：

        TF_REATTACH_PROVIDERS='{"fastly/fastly":{"Protocol":"grpc","Pid":54132,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/qm/swg2hf4h5t8sdht8yhds4dg6m0000gn/T/plugin865249851"}}}'
```

这也可通过两个独立的步骤来完成。`-gcflags` 选项用于禁用优化（`-N`）和内联编译（`-l`）。

```sh
$ go build -gcflags="all=-N -l" -o terraform-provider-fastly_debug
$ dlv exec terraform-provider-fastly_debug -- --debug
```

按照提示，在另一个终端中设置 `TF_REATTACH_PROVIDERS` 环境变量。之后像平常一样使用 Terraform，它就会自动在调试器中使用该提供程序。

```sh
$ export TF_REATTACH_PROVIDERS='{"fastly/fastly":{"Protocol":"grpc","Pid":54132,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/qm/swg2hf4h5t8sdht8yhds4dg6m0000gn/T/plugin865249851"}}}'
$ terraform plan
```

之后你就可以像平常一样使用调试器设置断点，并追踪该提供程序的执行流程了。

设置调试模式的实现方案假设使用的是 Terraform 0.13.x 版本。如果使用的是 Terraform 0.12.x，则需要手动修改分配给 `TF_REATTACH_PROVIDERS` 的值，将键 `"fastly/fastly"` 更改为 `"registry.terraform.io/-/fastly"`。更多详情可参阅 HashiCorp 的[“可调试提供程序二进制文件的支持”](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#support-for-debuggable-provider-binaries)文档。

### 摘要

```
（第一个shell）dlv debug. --headless -- --debug
（第二个shell）dlv connect <第一个shell的输出>
               继续
               <Ctrl-c>
               在 fastly/block_fastly_service_package.go:123 处中断
（第三个shell）export TF_REATTACH_PROVIDERS="..."
               terraform apply
（第二个shell）继续（进行逐步调试）
               <Ctrl-c>（然后在第三个shell中运行另一个terraform命令）
```

## 测试

要测试该提供程序，只需运行 `make test` 即可。

```sh
$ make test
```

要运行完整的验收测试套件，请执行 `make testacc`。

*注意：* 接受测试会创建真实的资源，且运行往往需要花费费用。因此完整套接受测试的运行时间可能会长达数小时。

```sh
$ make testacc
```

若要运行某一项验收测试，可结合正则表达式使用‘-run’标志。
以下示例使用了正则表达式来匹配名为‘TestAccFastlyServiceVCL_basic’的单一测试。

```sh
$ make testacc TESTARGS='-run=TestAccFastlyServiceVCL_basic'
```

以下示例使用正则表达式来执行一组基础验收测试。

```sh
$ make testacc TESTARGS='-run=TestAccFastlyServiceVCL.*_basic'
```

若要在运行测试时添加额外的调试信息，可在 `make` 命令前加上 `TF_LOG` 参数（详情请参阅 [terraform 文档](https://www.terraform.io/docs/internals/debugging.html)）。

```sh
$ TF_LOG=trace make testacc
```

默认情况下，测试以4个并行任务的方式运行。如果某些测试因网络问题而失败，可降低并行度；若有可能，则可提高并行度以缩短测试运行时间。可通过在`make`命令前加上`TEST_PARALLELISM`参数来配置此项，如下例所示。

```sh
$ TEST_PARALLELISM=8 make testacc
```

根据所使用的 Fastly 账户不同，某些功能可能无法启用（例如 Platform TLS）。在运行完整测试套件时，这可能会导致部分测试失败，并可能出现 `403 Unauthorised` 错误。请查阅 [Fastly API 文档](https://developer.fastly.com/reference/api/)，确认这些失败的测试是否使用了处于有限可用状态的功能，或是仅对特定客户开放的功能。如果是这种情况，要么使用上文所述的 `TESTARGS` 正则表达式，要么在需要排除的测试文件顶部临时添加 `t.SkipNow()`。

## 构建文档

请参阅[文档指南](./DOCUMENTATION.md)。

## 贡献指南

请参考 [CONTRIBUTING.md](./CONTRIBUTING.md) 文件。
