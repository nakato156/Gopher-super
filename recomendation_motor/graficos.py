import matplotlib.pyplot as plt

# Datos
workers = [6, 12, 18, 24, 48, 72]
tiempo_total = [0.290, 0.293, 0.304, 0.320, 0.362, 0.388]  # en segundos
speedup = [1.00, 0.99, 0.95, 0.91, 0.80, 0.75]
eficiencia = [100, 49.4, 31.7, 22.8, 10.0, 6.2]

# Crear figura con 3 subplots
fig, axes = plt.subplots(1, 3, figsize=(18,5))

# -----------------------------
# 1. Speedup vs Workers
axes[0].plot(workers, speedup, marker='o', linestyle='-', color='tab:blue')
axes[0].set_title('Speedup vs Workers')
axes[0].set_xlabel('Workers')
axes[0].set_ylabel('Speedup')
axes[0].grid(True)
axes[0].set_xticks(workers)
# Etiquetas de valores
for x, y in zip(workers, speedup):
    axes[0].text(x, y + 0.02, f'{y:.2f}', ha='center')

# -----------------------------
# 2. Eficiencia vs Workers
axes[1].plot(workers, eficiencia, marker='o', linestyle='-', color='tab:green')
axes[1].set_title('Eficiencia vs Workers')
axes[1].set_xlabel('Workers')
axes[1].set_ylabel('Eficiencia (%)')
axes[1].grid(True)
axes[1].set_xticks(workers)
for x, y in zip(workers, eficiencia):
    axes[1].text(x, y + 2, f'{y:.1f}%', ha='center')

# -----------------------------
# 3. Tiempo Total vs Workers
axes[2].plot(workers, tiempo_total, marker='o', linestyle='-', color='tab:red')
axes[2].set_title('Tiempo Total vs Workers')
axes[2].set_xlabel('Workers')
axes[2].set_ylabel('Tiempo Total (s)')
axes[2].grid(True)
axes[2].set_xticks(workers)
for x, y in zip(workers, tiempo_total):
    axes[2].text(x, y + 0.005, f'{y:.3f}', ha='center')

# Ajustes finales
plt.tight_layout()
plt.savefig('analisis_workers.png', dpi=300)  # Guardar imagen para README
plt.show()