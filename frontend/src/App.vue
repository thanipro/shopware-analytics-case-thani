<template>
  <div class="container">
    <header>
      <h1>Shopware Analytics Dashboard</h1>
    </header>

    <div v-if="loading" class="loading">Loading analytics data...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else class="dashboard">
      <div class="card">
        <h3>Page Views</h3>
        <p class="metric">{{ analytics.total_page_views }}</p>
      </div>

      <div class="card">
        <h3>Add to Carts</h3>
        <p class="metric">{{ analytics.total_add_to_carts }}</p>
      </div>

      <div class="card">
        <h3>Purchases</h3>
        <p class="metric">{{ analytics.total_purchases }}</p>
      </div>

      <div class="card">
        <h3>Conversion Rate</h3>
        <p class="metric">{{ analytics.conversion_rate }}%</p>
      </div>

      <div class="card">
        <h3>Average Order Value</h3>
        <p class="metric">${{ analytics.average_purchase_value }}</p>
      </div>

      <div class="card">
        <h3>Max Order Value</h3>
        <p class="metric">${{ analytics.max_purchase_value }}</p>
      </div>

      <div class="card">
        <h3>Min Order Value</h3>
        <p class="metric">${{ analytics.min_purchase_value }}</p>
      </div>

      <div class="card">
        <h3>Top Product</h3>
        <p class="metric">{{ analytics.top_product_id || 'N/A' }}</p>
      </div>
    </div>

    <footer>
      <p>Last updated: {{ lastUpdated }}</p>
      <p>Auto-refreshes every 5 seconds</p>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import type { Analytics } from './types'

const analytics = ref<Analytics>({
  total_page_views: 0,
  total_add_to_carts: 0,
  total_purchases: 0,
  conversion_rate: 0.0,
  average_purchase_value: 0.0,
  max_purchase_value: 0.0,
  min_purchase_value: 0.0,
  top_product_id: null
})

const loading = ref<boolean>(true)
const error = ref<string | null>(null)
const lastUpdated = ref<string | null>(null)
let intervalId: number | undefined

const fetchAnalytics = async (): Promise<void> => {
  try {
    const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8000'
    const response = await fetch(`${apiUrl}/api/analytics`)

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    analytics.value = await response.json()
    lastUpdated.value = new Date().toLocaleTimeString()
    loading.value = false
    error.value = null
  } catch (e) {
    error.value = `Failed to fetch analytics: ${(e as Error).message}`
    loading.value = false
  }
}

onMounted(async () => {
  await fetchAnalytics()
  intervalId = setInterval(() => {
    fetchAnalytics()
  }, 5000)
})

onBeforeUnmount(() => {
  if (intervalId) {
    clearInterval(intervalId)
  }
})
</script>

<style scoped>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
}

header {
  text-align: center;
  margin-bottom: 40px;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 10px;
}

header h1 {
  font-size: 2.5em;
  font-weight: 600;
}

.loading,
.error {
  text-align: center;
  padding: 40px;
  font-size: 1.2em;
}

.error {
  color: #e53e3e;
  background: #fff5f5;
  border-radius: 8px;
  border: 1px solid #fc8181;
}

.dashboard {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 40px;
}

.card {
  background: white;
  padding: 25px;
  border-radius: 10px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s, box-shadow 0.2s;
}

.card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 12px rgba(0, 0, 0, 0.15);
}

.card h3 {
  font-size: 0.9em;
  color: #718096;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
}

.metric {
  font-size: 2.5em;
  font-weight: 700;
  color: #2d3748;
}

footer {
  text-align: center;
  padding: 20px;
  color: #718096;
  font-size: 0.9em;
}

footer p {
  margin: 5px 0;
}
</style>
