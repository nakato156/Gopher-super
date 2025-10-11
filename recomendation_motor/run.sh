#!/usr/bin/env bash
set -euo pipefail

# ---------------------------
# Config rápida
# ---------------------------
APP_BIN="./bin/recom"
CONFIG_FILE="config.json"
PERSIST_DIR="data/persistence"

# Modo por defecto:
#   pipeline  -> PASOS 1-4 (NO corre compare)
#   compare   -> SOLO PASO 5 (re-ejecuta internamente secuencial+concurrente)
#   concurrent|sequential -> sólo ese motor
MODE="${1:-pipeline}"

# ---------------------------
# Helpers
# ---------------------------
banner() { echo -e "\n$1\n"; }
section() {
  echo ""
  echo "═══════════════════════════════════════════════════════════════"
  echo "$1"
  echo "═══════════════════════════════════════════════════════════════"
  echo ""
}
ok()   { echo "✔ $1"; }
fail() { echo "✘ $1"; exit 1; }

trap 'echo "Error en línea ${LINENO}. Abortando."; exit 1' ERR

# ---------------------------
# 0) Verificar/crear config
# ---------------------------
banner "╔════════════════════════════════════════════════════════════╗
║   PIPELINE COMPLETO DE RECOMENDACIONES                     ║
║   CON MÉTRICAS DE RENDIMIENTO Y CONFIGURACIÓN              ║
╚════════════════════════════════════════════════════════════╝"

echo "Verificando configuración..."
if [ ! -f "$CONFIG_FILE" ]; then
  cat > "$CONFIG_FILE" << 'EOF'
{
  "concurrency": {
    "parser_workers": 6,
    "similarity_workers": 12,
    "recommendation_workers": 6,
    "buffer_size": 2000
  },
  "sampling": {
    "percentage": 10,
    "random_seed": 42
  },
  "min_common_games": 1,
  "min_similarity_score": 0.1,
  "k": 10,
  "n": 5
}
EOF
  ok "Configuración creada en $CONFIG_FILE"
else
  echo "Configuración encontrada. Mostrando:"
  cat "$CONFIG_FILE"
fi

mkdir -p "$PERSIST_DIR"

# ---------------------------
# 1) Build único (rápido)
# ---------------------------
section "BUILD"
if [ ! -x "$APP_BIN" ]; then
  mkdir -p "$(dirname "$APP_BIN")"
  go build -o "$APP_BIN" .
  ok "Binario compilado en $APP_BIN"
else
  # Rebuild si fuentes son más nuevas que binario
  if [ "$(find . -name '*.go' -newer "$APP_BIN" | wc -l)" -gt 0 ]; then
    go build -o "$APP_BIN" .
    ok "Binario actualizado en $APP_BIN"
  else
    ok "Binario ya está al día"
  fi
fi

# ---------------------------
# Funciones de pasos
# ---------------------------
sampling_and_clean() {
  section "PASO 1: SAMPLING Y LIMPIEZA DE COLUMNAS"
  if [ ! -f "steam_reviews_sample_10pct.csv" ]; then
    echo "Ejecutando sampling y limpieza…"
    "$APP_BIN" cleaner --config="$CONFIG_FILE"
    ok "Sampling y limpieza"
  else
    echo "Archivo de muestra ya existe, saltando"
  fi
}

parse_step() {
  section "PASO 2: PARSING DE DATOS"
  if [ ! -f "$PERSIST_DIR/user_profiles_sample.gob" ] || [ ! -f "$PERSIST_DIR/game_names_sample.gob" ]; then
    echo "Ejecutando parser…"
    "$APP_BIN" parser --config="$CONFIG_FILE"
    ok "Parsing de datos"
  else
    echo "Persistencia ya existe, saltando parser"
  fi
}

motor_concurrent() {
  section "PASO 3: MOTOR DE RECOMENDACIONES (CONCURRENTE)"
  "$APP_BIN" motor --config="$CONFIG_FILE"
  ok "Motor concurrente"
}

motor_sequential() {
  section "PASO 4: MOTOR DE RECOMENDACIONES (SECUENCIAL)"
  "$APP_BIN" sequential --config="$CONFIG_FILE"
  ok "Motor secuencial"
}

compare_step() {
  section "PASO 5: COMPARACIÓN DE RENDIMIENTO"
  "$APP_BIN" compare --config="$CONFIG_FILE"
  ok "Comparación de rendimiento"
}

# ---------------------------
# Ejecución por modo
# ---------------------------
case "$MODE" in
  pipeline)
    banner "INICIANDO PIPELINE (sin comparación)"
    sampling_and_clean
    parse_step
    motor_concurrent
    motor_sequential
    ;;
  compare)
    banner "INICIANDO COMPARACIÓN SOLAMENTE"
    # No ejecutamos motor/sequential antes: compare ya los corre internamente.
    # Aun así, aseguramos que existan datos base:
    sampling_and_clean
    parse_step
    compare_step
    ;;
  concurrent)
    banner "EJECUTANDO SÓLO CONCURRENTE"
    sampling_and_clean
    parse_step
    motor_concurrent
    ;;
  sequential)
    banner "EJECUTANDO SÓLO SECUENCIAL"
    sampling_and_clean
    parse_step
    motor_sequential
    ;;
  *)
    fail "Modo inválido: $MODE (usa: pipeline | compare | concurrent | sequential)"
    ;;
esac

echo -e "\nListo."
