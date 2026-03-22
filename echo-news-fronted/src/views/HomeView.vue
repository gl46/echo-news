<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { fetchArticles } from '@/api'
import type { ArticleSummary } from '@/api/types'

const router = useRouter()
const articles = ref<ArticleSummary[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    articles.value = await fetchArticles()
  } catch (e: any) {
    error.value = e.message ?? 'Failed to load articles'
  } finally {
    loading.value = false
  }
})

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-GB', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function goToArticle(id: string) {
  router.push(`/articles/${id}`)
}
</script>

<template>
  <div class="home">
    <h1>EchoNews</h1>
    <p class="subtitle">BBC News - Bilingual Reader</p>

    <div v-if="loading" class="status">Loading articles...</div>
    <div v-else-if="error" class="status error">{{ error }}</div>
    <div v-else class="article-list">
      <div
        v-for="article in articles"
        :key="article.id"
        class="article-card"
        @click="goToArticle(article.id)"
      >
        <div class="article-meta">
          <span class="source">{{ article.source }}</span>
          <span class="date">{{ formatDate(article.published_at) }}</span>
          <span v-if="article.has_translation" class="translated-badge">Translated</span>
        </div>
        <h2 class="article-title">{{ article.title }}</h2>
        <p class="article-summary">{{ article.summary }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home {
  max-width: 720px;
  margin: 0 auto;
  padding: 24px 16px;
}

h1 {
  font-size: 28px;
  margin: 0;
}

.subtitle {
  color: #666;
  margin: 4px 0 24px;
}

.status {
  text-align: center;
  padding: 40px;
  color: #666;
}

.status.error {
  color: #d32f2f;
}

.article-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.article-card {
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: box-shadow 0.2s;
}

.article-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.article-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  margin-bottom: 8px;
}

.source {
  background: #1a73e8;
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
  font-size: 12px;
}

.date {
  color: #888;
}

.translated-badge {
  background: #4caf50;
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.article-title {
  font-size: 18px;
  margin: 0 0 8px;
  line-height: 1.4;
}

.article-summary {
  font-size: 14px;
  color: #555;
  margin: 0;
  line-height: 1.5;
}
</style>
