#!/bin/bash

echo "╔════════════════════════════════════════════════════════════╗"
echo "║   PIPELINE COMPLETO DE RECOMENDACIONES                      ║"
echo "║   CON MÉTRICAS DE RENDIMIENTO Y CONFIGURACIÓN              ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Función para mostrar el progreso
show_progress() {
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "🔄 $1"
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
}

# Función para verificar si un comando fue exitoso
check_success() {
    if [ $? -eq 0 ]; then
        echo "✅ $1 completado exitosamente"
    else
        echo "❌ Error en $1"
        exit 1
    fi
}

echo "🔧 Verificando configuración..."
if [ ! -f "config.json" ]; then
    echo "📁 Creando archivo de configuración por defecto..."
    cat > config.json << EOF
{
  "concurrency": {
    "parser_workers": 4,
    "similarity_workers": 8,
    "recommendation_workers": 4,
    "buffer_size": 1000
  },
  "sampling": {
    "percentage": 10,
    "random_seed": 42
  },
  "min_common_games": 1,
  "min_similarity_score": 0.01,
  "k": 10,
  "n": 5
}
EOF
    echo "✅ Configuración creada"
else
    echo "✅ Configuración encontrada"
    echo "📋 Mostrando configuración actual:"
    cat config.json
fi

echo ""
echo "🚀 INICIANDO PIPELINE COMPLETO"
echo ""

# PASO 1: SAMPLING Y LIMPIEZA DE COLUMNAS (solo si no existe el archivo de muestra)
show_progress "PASO 1: SAMPLING Y LIMPIEZA DE COLUMNAS"
if [ ! -f "steam_reviews_sample_10pct.csv" ]; then
    echo "📊 Archivo de muestra no encontrado, ejecutando sampling y limpieza..."
    go run . cleaner
    check_success "Sampling y limpieza"
else
    echo "✅ Archivo de muestra ya existe, saltando sampling y limpieza"
fi

# PASO 2: PARSING (solo si no existen archivos de persistencia)
show_progress "PASO 2: PARSING DE DATOS"
if [ ! -f "data/persistence/user_profiles_sample.gob" ] || [ ! -f "data/persistence/game_names_sample.gob" ]; then
    echo "📊 Archivos de persistencia no encontrados, ejecutando parser..."
    go run . parser
    check_success "Parsing de datos"
else
    echo "✅ Archivos de persistencia ya existen, saltando parser"
fi

# PASO 3: MOTOR DE RECOMENDACIONES
show_progress "PASO 3: MOTOR DE RECOMENDACIONES"
echo "🎯 Generando recomendaciones con análisis de rendimiento..."
go run . motor
check_success "Motor de recomendaciones"

# RESUMEN FINAL
echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║                    PIPELINE COMPLETADO                     ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "✅ TODOS LOS PASOS COMPLETADOS EXITOSAMENTE:"
echo "   1️⃣ Sampling y limpieza de columnas"
echo "   2️⃣ Parsing de datos (si era necesario)"
echo "   3️⃣ Motor de recomendaciones"
echo ""
echo "📊 MÉTRICAS INCLUIDAS EN CADA PASO:"
echo "   - Tiempo de cómputo (ms y segundos)"
echo "   - Elementos procesados por segundo"
echo "   - Número de workers utilizados"
echo "   - Análisis de paralelización"
echo ""
echo "💡 Para modificar la configuración, edita el archivo config.json"
echo "💡 Para ejecutar solo un paso específico:"
echo "   - go run . cleaner  (sampling y limpieza)"
echo "   - go run . parser   (parsing de datos)"
echo "   - go run . motor    (motor de recomendaciones)"
