# 📊 Análisis de Datos – MovieLens 32M (Programación Concurrente y Distribuida)

Este módulo en **Go** realiza el **análisis, limpieza y normalización** de los datos del dataset **MovieLens 32M**, correspondiente a la **Etapa 1: Análisis de Datos** del trabajo final del curso *Programación Concurrente y Distribuida (UPC, 2025-2)*.

---

##  Funcionalidades principales

- **Lectura y validación** de los datos del archivo `ratings.csv`.  
- **Verificación de valores nulos o incompletos**.  
- **Generación de la matriz usuario–película** en memoria.  
- **Normalización de ratings** por usuario (centrando cada calificación según el promedio individual).  
- **Exportación de la matriz resultante** a formato `matriz_normalizada.json`, lista para usar en el modelo de **filtrado colaborativo** (item-based).  

---

##  Estructura de archivos

```
movie_lens_data_procc/
│
├── analisis.go                 # Código principal en Go
├── matriz_normalizada.json     # Archivo generado (no se sube a GitHub por su tamaño)
├── ml-32m/                     # Carpeta local con los CSV del dataset MovieLens
│   ├── ratings.csv
│   ├── movies.csv
│   └── tags.csv
└── README.md                   # Este archivo
```

> ⚠️ Nota: por el tamaño del dataset, el archivo `matriz_normalizada.json` **no se incluye en el repositorio**.  
> Puede regenerarse ejecutando el programa con el dataset original.

---

##  Ejecución

1. Asegúrate de tener **Go 1.25** instalado.  
2. Descarga el dataset **MovieLens 32M** y colócalo en la carpeta `ml-32m/`.  
3. Ejecuta el script desde la terminal:

   ```bash
   go run analisis.go
   ```

---

##  Salida esperada

El programa imprimirá en consola mensajes de progreso como:

```
📥 Leyendo archivo ratings.csv ...
Total de registros leídos: 32000204
✅ No se encontraron valores nulos o incompletos.
✅ Matriz usuario–película creada correctamente.
Usuarios totales: 200948
✅ Normalización completada.
💾 Matriz guardada en 'matriz_normalizada.json'
Usuario 43972 → map[318:0.48 527:0.98 858:0.48 ...]
Usuario 48616 → map[1:0.49 296:0.49 2959:0.99 ...]
Promedio global normalizado: -0.0000, Desviación estándar: 0.9393
```

Estos valores muestran las calificaciones **normalizadas** de algunos usuarios seleccionados.  
Al final, se muestra también la media y desviación estándar global del conjunto de datos.
