#!/bin/bash

set -e

cd "$(dirname "${BASH_SOURCE[0]}")"

BIN_NAME="tick"

# Options
BUILD=false
INSTALL=false
HELP=false


# Parsear opciones
while [[ "$#" -gt 0 ]]; do
    case "$1" in
        --h | --help) HELP=true ;;                    # Activar la bandera si se proporciona --help
        --b | --build) BUILD=true ;;                # Activar la bandera si se proporciona --deploy
        --i | --install) INSTALL=true ;;                    # Activar la bandera si se proporciona --undo
        *) echo "Opción desconocida: $1"; exit 1 ;;
    esac
    shift
done

help() {
  echo ""
  echo "Uso: command.sh [opciones]"
  echo "Opciones:"
  echo "  --h, --help: Muestra este mensaje de ayuda."
  echo "  --b, --build: Compila el binario en este directorio."
  echo "  --i, --install: Compila e instala el binario en /usr/local/bin."
  echo ""
  exit 0
}

# build the project
build() {
  echo "Compilando $BIN_NAME..."
  go build -o "$BIN_NAME" ./cmd
  echo "Listo: ./$BIN_NAME"
}

# install the project
install() {
  echo "Compilando $BIN_NAME..."
  go build -o "$BIN_NAME" ./cmd

  echo "Instalando $BIN_NAME en /usr/local/bin..."
  sudo cp "$BIN_NAME" /usr/local/bin/
  sudo chmod +x /usr/local/bin/"$BIN_NAME"
  echo "Listo: /usr/local/bin/$BIN_NAME"
}

# Main
if [ "$HELP" = true ]; then
  help
elif [ "$BUILD" = true ]; then
  build
elif [ "$INSTALL" = true ]; then
  install
else
  help
fi

# Exit with success
exit 0