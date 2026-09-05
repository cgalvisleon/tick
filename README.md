# tick

CLI en Go para llevar seguimiento simple de proyectos y tareas, con una ergonomía inspirada en git: cada proyecto vive en un directorio `.tick/` (como `.git/`) que contiene una base de datos SQLite (`tick.db`), y se sincroniza entre máquinas con `remote`, `push` y `pull`.

## Requisitos

- Go 1.25 o superior
- Library Et:
  go get github.com/cgalvisleon/et@v1.0.32

## Instalación / ejecución

Clona o ubícate en este directorio y compila el binario:

```sh
go build -o tick ./cmd
```

Esto genera un ejecutable `tick` en el directorio actual. Puedes moverlo a un directorio en tu `PATH` (por ejemplo `~/go/bin` o `/usr/local/bin`) para usarlo desde cualquier lugar:

```sh
mv tick ~/go/bin/
```

También puedes ejecutarlo sin compilar, usando `go run ./cmd` (útil durante desarrollo):

```sh
go run ./cmd init
```

En el resto de este documento se usa `tick` asumiendo que el binario está en el `PATH`; si no lo está, sustituye por `go run ./cmd`.

### Script command.sh

También se incluye `command.sh` como atajo para compilar/instalar:

```sh
./command.sh --build    # o --b   -> compila ./tick
./command.sh --install  # o --i   -> compila e instala en /usr/local/bin
./command.sh --help     # o --h   -> muestra la ayuda
```

## Uso rápido

```sh
# Inicializar un proyecto en el directorio actual
tick init

# Configuración del proyecto (guardada en .tick/tick.db, no es global)
tick config user.name "Cesar Galvis"
tick config user.email "cgalvisleon@gmail.com"

# Ver / actualizar los datos del proyecto
tick project
tick project code:P1 name:"Mi proyecto" description:"Descripcion del proyecto"
tick project tag prioridad alta
tick project tag remove prioridad

# Crear y listar tareas
tick task code:T1 name:"Primera tarea" type:feature assignee:cesar
tick task
tick task code:T1

# Actualizar campos de una tarea existente
tick task code:T1 name:"Nuevo nombre"

# Registrar avances de estado (pending | in_process | stop | await | done)
tick status code:T1 status:in_process description:"Empezando" percent:0
tick status code:T1 status:done description:"Terminado" percent:100

# Ver historial de estados de una tarea
tick task code:T1 status

# Tags sobre una tarea
tick task code:T1 tag sprint 3
tick task code:T1 tag remove sprint

# Configurar un remote (ruta del sistema de archivos) y sincronizar
tick remote add origin /ruta/compartida/mi-proyecto
tick remote
tick push          # copia la base local hacia el remote (sobreescribe el remote)
tick pull          # copia la base del remote hacia lo local (sobreescribe lo local)
```

## Comandos

Todos los comandos (excepto `init`) buscan el directorio `.tick/` más cercano subiendo desde el directorio actual, igual que git busca `.git/`.

Los argumentos de "campo:valor" (`code:T1`, `status:done`, etc.) se pueden combinar libremente; las palabras sueltas sin `:` (como `tag` o `status`) actúan como subcomandos.

| Comando                                                                              | Descripción                                                                                                            |
| ------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------- |
| `tick init`                                                                          | Inicializa un proyecto tick en el directorio actual (`.tick/`)                                                         |
| `tick config <key> [value]`                                                          | Lee o escribe la configuración del proyecto (guardada en `.tick/tick.db`), p. ej. `user.name` / `user.email` / `token` |
| `tick project [campo:valor ...]`                                                     | Muestra o actualiza los datos del proyecto actual                                                                      |
| `tick project tag <nombre> <valor>` / `tag remove <nombre>`                          | Agrega, actualiza o elimina un tag del proyecto                                                                        |
| `tick task [ID:id\|code:codigo] [campo:valor ...]`                                   | Sin argumentos, lista las tareas; con identificador, crea o actualiza una tarea                                        |
| `tick task ID:id\|code:codigo status`                                                | Muestra el historial de estados de una tarea                                                                           |
| `tick task ID:id\|code:codigo tag <nombre> <valor>` / `tag remove <nombre>`          | Agrega, actualiza o elimina un tag de una tarea                                                                        |
| `tick status ID:id\|code:codigo status:<estado> description:<texto> percent:<0-100>` | Registra un avance de estado para una tarea                                                                            |
| `tick remote`                                                                        | Lista los remotes configurados                                                                                         |
| `tick remote add <nombre> <path>`                                                    | Agrega o actualiza un remote (ruta del sistema de archivos)                                                            |
| `tick remote remove <nombre>`                                                        | Elimina un remote                                                                                                      |
| `tick push [remote]`                                                                 | Copia la base de datos local hacia el remote (por defecto `origin`); sobreescribe el remote por completo               |
| `tick pull [remote]`                                                                 | Copia la base de datos del remote hacia el proyecto local (por defecto `origin`); sobreescribe lo local por completo   |

### Estados de una tarea

Los estados válidos son: `pending`, `in_process`, `stop`, `await`, `done` (también se aceptan variantes como `in process` o `in-process`). El tiempo real de una tarea (`actual_minutes`) se calcula desde que entra por primera vez a `in_process` hasta que pasa a `done`, descontando el tiempo que estuvo en `stop` o `await`.

### push / pull

`push` y `pull` **no hacen merge**: son una copia completa del archivo `tick.db` entre el proyecto local y la ruta del remote, sobreescribiendo el destino por completo. Úsalos con un remote que apunte a una carpeta compartida (por ejemplo una unidad de red o una carpeta sincronizada), no como control de versiones.

## Estructura del proyecto

```
cmd/         Punto de entrada (main.go)
pkg/tick/    Comandos de la CLI (Cobra)
internal/
  findroot/  Localiza el directorio .tick/ más cercano
  store/     Capa de persistencia (SQLite vía et/jsql): Project, Task, Remote, Config
  ui/        Salida coloreada de estados y barras de porcentaje
command.sh   Script de build/instalación
```

Para más detalle de la arquitectura interna, ver [CLAUDE.md](CLAUDE.md).

## Install

```
sudo cp tick /usr/local/bin/
sudo chmod +x /usr/local/bin/tick
```
