# Proveedor Terraform de Fastly

<!-- hy-mt2-i18n:start -->
[English](./README.md) | [中文](./README_zh-CN.md) | [日本語](./README_ja.md) | **Español**
<!-- hy-mt2-i18n:end -->


- Página web: https://www.terraform.io  
- Documentación: https://registry.terraform.io/providers/fastly/fastly/latest/docs  
- Problemas reportados: https://github.com/fastly/terraform-provider-fastly/blob/main/ISSUES.md  
- Lista de correo: http://groups.google.com/group/terraform-tool  
- [![Chat en Gitter](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)

## Requisitos

- [Terraform](https://www.terraform.io/downloads.html) 0.12.x o superior  
- [Go](https://golang.org/doc/install) 1.25 (para compilar el complemento del proveedor)

NOTA: la última versión del proveedor Fastly que soportaba Terraform 0.11.x y versiones anteriores fue [v0.26.0](https://github.com/fastly/terraform-provider-fastly/releases/tag/v0.26.0).

## Versionado y calendarios de lanzamiento

Los mantenedores de este módulo se esfuerzan por seguir el [versionado semántico (SemVer)](https://semver.org/). Esto implica que los cambios que rompen la compatibilidad (eliminación de funcionalidades o modificaciones incompatibles en las existentes) se publicarán en una versión en la que se incrementa el primer componente de la versión (`major`). Las adiciones de nuevas funcionalidades harán aumentar el segundo componente de la versión (`minor`), mientras que las correcciones de errores que no afectan la compatibilidad incrementarán el tercer componente de la versión (`patch`).

El tercer miércoles de cada mes se publicará una nueva versión que incluirá todos los cambios relacionados con modificaciones significativas, nuevas funcionalidades y correcciones de errores listos para su lanzamiento. Si ese miércoles coincide con un día festivo en Estados Unidos, el lanzamiento se pospondrá hasta el próximo día hábil disponible.

Si hay correcciones de errores críticas o urgentes listas para ser publicadas entre esas versiones principales, se lanzarán versiones de parche según sea necesario para poner esas correcciones a disposición.

## Compilación del proveedor

Clonar el repositorio en: `$GOPATH/src/github.com/fastly/terraform-provider-fastly`

```sh
$ mkdir -p $GOPATH/src/github.com/fastly; cd $GOPATH/src/github.com/fastly
$ git clone git@github.com:fastly/terraform-provider-fastly
```

Ingrese al directorio del proveedor e compile el mismo.

```sh
$ cd $GOPATH/src/github.com/fastly/terraform-provider-fastly
$ make build
```

## Desarrollo del proveedor

Si desea trabajar en el proveedor, primero necesitará tener instalado [Go](http://www.golang.org) en su equipo (se *requiere* la versión 1.25 o superior).

Para compilar el proveedor, ejecute `make build`. Esto generará el proveedor y colocará su binario en una carpeta local llamada `bin`.

```sh
$ make build
...
```

Junto con el binario recién generado, se creará un archivo llamado `developer_overrides.tfrc`. El objetivo `make build` comunicará los detalles necesarios para establecer la variable de entorno `TF_CLI_CONFIG_FILE`, lo que permitirá a Terraform utilizar el binario del proveedor compilado localmente.

- HashiCorp - [Sobrescrituras de desarrollo para desarrolladores de proveedores](https://www.terraform.io/docs/cli/config/config-file.html#development-overrides-for-provider-developers).

> **NOTA**: Si tiene problemas para observar los efectos de los cambios en el código del proveedor, es posible que la CLI de Terraform no sepa qué binario del proveedor debe utilizar. Revise el directorio `./bin/` para comprobar si existen varios proveedores con diferentes hashes de commit (por ejemplo, `terraform-provider-fastly_v2.2.0-5-gfdc37cee`) y elimínelos primero antes de ejecutar `make build`. Esto debería ayudar a que la CLI de Terraform elija el binario correcto.

### Depuración del proveedor

El método anterior con `dev_overrides` debería ser suficiente para la mayoría de los usos en entornos de desarrollo, incluyendo probar los cambios locales con código de Terraform real. Sin embargo, a veces puede resultar útil ejecutar el proveedor en modo de depuración si es necesario conectar un depurador como [delve](https://github.com/go-delve/delve) para solucionar un problema específico.

Por defecto, Terraform inicia al proveedor en un subproceso y se conecta a él mediante GRPC a través de un socket local.  
(Para más información al respecto, consulte [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin#architecture)).  
Para depurar, es posible omitir este proceso e ejecutar al proveedor en un proceso separado, indicándole luego a Terraform cómo comunicarse con él.  
La ventaja de esto es que el proveedor puede iniciarse con un depurador conectado, de la misma manera en que el depurador se conectaría a cualquier ejecutable normal.

Existen varias formas de hacerlo según el depurador que se utilice, pero aquí emplearemos [delve](https://github.com/go-delve/delve), ya que el proceso debería ser bastante similar con otros depuradores. Los dos aspectos que deben configurarse son compilar el proveedor sin optimizaciones y ejecutar el archivo ejecutable con la opción `--debug`. Compilar sin optimizaciones garantiza que el depurador pueda acceder a todos los símbolos del binario que necesita, mientras que la opción `--debug` indica al SDK del plugin de Terraform que debe ejecutarse en un proceso separado y mostrar las instrucciones para conectarse a él una vez iniciado.

Con [delve](https://github.com/go-delve/delve) esto se puede hacer con una sola orden:

```sh
$ dlv debug. -- --debug
Escriba ‘help’ para ver la lista de comandos.
(dlv) continue
{"@level":"debug","@message":"dirección del plugin","@timestamp":"2021-03-26T12:10:13.320981Z","address":"/var/folders/qm/swg2hf4h5t8sdht8yhds4dg6m0000gn/T/plugin865249851","network":"unix"}
El proveedor se ha iniciado; para conectarse a Terraform establezca la variable de entorno TF_REATTACH_PROVIDERS:

        TF_REATTACH_PROVIDERS='{"fastly/fastly":{"Protocol":"grpc","Pid":54132,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/qm/swg2hf4h5t8sdht8yhds4dg6m0000gn/T/plugin865249851"}}}'
```

Esto también se puede realizar en dos pasos separados. La opción `-gcflags` desactiva las optimizaciones (`-N`) y la inclusión de código (`-l`).

```sh
$ go build -gcflags="all=-N -l" -o terraform-provider-fastly_debug
$ dlv exec terraform-provider-fastly_debug -- --debug
```

Como indica el mensaje, vaya a otra shell e importe la variable de entorno `TF_REATTACH_PROVIDERS`. A continuación, utilice Terraform como de costumbre, y este utilizará automáticamente al proveedor a través del depurador.

```sh
$ export TF_REATTACH_PROVIDERS='{"fastly/fastly":{"Protocol":"grpc","Pid":54132,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/qm/swg2hf4h5t8sdht8yhds4dg6m0000gn/T/plugin865249851"}}}'
$ terraform plan
```

A continuación, podrá establecer puntos de interrupción y rastrear la ejecución del proveedor mediante el depurador, tal como es de esperar.

La implementación para configurar el modo de depuración presupone que se está utilizando Terraform 0.13.x. Si utiliza Terraform 0.12.x, deberá modificar manualmente el valor asignado a `TF_REATTACH_PROVIDERS` de modo que la clave `"fastly/fastly"` pase a ser `"registry.terraform.io/-/fastly"`. Consulte la sección de [HashiCorp sobre el soporte para binarios de proveedores depurables](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#support-for-debuggable-provider-binaries) para obtener más detalles.

### Resumen

```
(primer shell)  dlv debug. --headless -- --debug
(segundo shell) dlv connect <resultado del primer shell>
               continuar
               <Ctrl-c>
               interrumpir en fastly/block_fastly_service_package.go:123
(tercer shell)  export TF_REATTACH_PROVIDERS="..."
               terraform apply
(segundo shell) continuar (realizar los pasos de depuración)
               <Ctrl-c> (luego ejecutar otra orden de terraform desde el tercer shell)
```

## Pruebas

Para probar el proveedor, basta con ejecutar `make test`.

```sh
$ make test
```

Para ejecutar el conjunto completo de pruebas de aceptación, ejecute `make testacc`.

*Nota:* Las pruebas de aceptación crean recursos reales y, con frecuencia, su ejecución implica gastos económicos. Debe tener en cuenta que la ejecución de todo el conjunto de pruebas de aceptación puede tardar horas.

```sh
$ make testacc
```

Para ejecutar una prueba de aceptación individual, se puede utilizar el parámetro ‘-run’ junto con una expresión regular.  
El ejemplo siguiente emplea una expresión regular que coincide con una prueba específica llamada ‘TestAccFastlyServiceVCL_basic’.

```sh
$ make testacc TESTARGS='-run=TestAccFastlyServiceVCL_basic'
```

El siguiente ejemplo utiliza una expresión regular para ejecutar un grupo de pruebas de aceptación básicas.

```sh
$ make testacc TESTARGS='-run=TestAccFastlyServiceVCL.*_basic'
```

Para ejecutar las pruebas con contexto de depuración adicional, añada `TF_LOG` al inicio de la orden `make` (consulte la [documentación de Terraform](https://www.terraform.io/docs/internals/debugging.html) para más detalles).

```sh
$ TF_LOG=trace make testacc
```

De forma predeterminada, las pruebas se ejecutan con una paralelización de 4. Esta cantidad puede reducirse si algunas pruebas fallan debido a problemas de red, o aumentarse en la medida de lo posible para acortar el tiempo de ejecución de las pruebas. Para configurarlo, añada el parámetro `TEST_PARALLELISM` al inicio de la orden `make`, como se muestra en el ejemplo siguiente.

```sh
$ TEST_PARALLELISM=8 make testacc
```

Dependiendo de la cuenta Fastly utilizada, es posible que algunas funcionalidades no estén habilitadas (por ejemplo, Platform TLS). Esto podría hacer que algunos tests fallen, con posibles errores de tipo `403 Unauthorised` al ejecutar el conjunto completo de pruebas. Consulte la [documentación de la API de Fastly](https://developer.fastly.com/reference/api/) para confirmar si los tests que fallan utilizan funcionalidades de disponibilidad limitada o que solo están disponibles para ciertos clientes. Si es así, utilice las expresiones regulares `TESTARGS` descritas anteriormente, o agregue temporalmente `t.SkipNow()` al principio de los tests que deben excluirse.

## Compilación de la documentación

Consulte la [guía de documentación](./DOCUMENTATION.md).

## Contribuir

Consulte [CONTRIBUTING.md](./CONTRIBUTING.md)
