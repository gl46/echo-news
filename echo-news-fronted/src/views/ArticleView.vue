<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchArticle, translateArticle } from '@/api'
import type { ArticleDetail } from '@/api/types'

const route = useRoute()
const router = useRouter()
const article = ref<ArticleDetail | null>(null)
const loading = ref(true)
const error = ref('')

type TranslateState = 'idle' | 'loading' | 'success' | 'error'
const translateState = ref<TranslateState>('idle')
const translateError = ref('')

onMounted(async () => {
  const id = route.params.id as string
  try {
    article.value = await fetchArticle(id)
    if (article.value.translation) {
      translateState.value = 'success'
    }
  } catch (e: any) {
    error.value = e.message ?? 'Failed to load article'
  } finally {
    loading.value = false
  }
})

async function handleTranslate() {
  if (!article.value || translateState.value === 'loading') return
  translateState.value = 'loading'
  translateError.value = ''
  try {
    const result = await translateArticle(article.value.id)
    article.value.translation = result.translation
    translateState.value = 'success'
  } catch (e: any) {
    translateError.value = e.message ?? 'Translation failed'
    translateState.value = 'error'
  }
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-GB', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div class="article-page">
    <button class="back-btn" @click="router.push('/')">&larr; Back</button>

    <div v-if="loading" class="status">Loading article...</div>
    <div v-else-if="error" class="status error">{{ error }}</div>
    <template v-else-if="article">
      <div class="article-meta">
        <span class="source">{{ article.source }}</span>
        <span class="date">{{ formatDate(article.published_at) }}</span>
      </div>

      <h1>{{ article.title }}</h1>

      <a :href="article.url" target="_blank" rel="noopener" class="original-link">
        View original article &rarr;
      </a>

      <section class="content-section">
        <h2>English</h2>
        <div class="content-body">
          <p v-for="(para, i) in article.content.split('\n\n')" :key="i">{{ para }}</p>
        </div>
      </section>

      <section class="content-section translation-section">
        <div class="translation-header">
          <h2>Translation</h2>
          <button
            class="translate-btn"
            :class="translateState"
            :disabled="translateState === 'loading' || translateState === 'success'"
            @click="handleTranslate"
          >
            <template v-if="translateState === 'idle'">Translate</template>
            <template v-else-if="translateState === 'loading'">Translating...</template>
            <template v-else-if="translateState === 'success'">Done</template>
            <template v-else-if="translateState === 'error'">Retry</template>
          </button>
        </div>
        <div v-if="translateState === 'error'" class="translate-error">
          {{ translateError }}
        </div>
        <div v-if="article.translation" class="content-body translation-body">
          <p v-for="(para, i) in article.translation.split('\n\n')" :key="i">{{ para }}</p>
        </div>
        <div v-else-if="translateState === 'idle'" class="translation-placeholder">
          Click "Translate" to get the Chinese translation.
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.article-page {
  max-width: 720px;
  margin: 0 auto;
  padding: 24px 16px;
}

.back-btn {
  background: none;
  border: 1px solid #ddd;
  border-radius: 6px;
  padding: 6px 14px;
  cursor: pointer;
  font-size: 14px;
  color: #333;
  margin-bottom: 20px;
}

.back-btn:hover {
  background: #f5f5f5;
}

.status {
  text-align: center;
  padding: 40px;
  color: #666;
}

.status.error {
  color: #d32f2f;
}

.article-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  margin-bottom: 12px;
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

h1 {
  font-size: 24px;
  line-height: 1.4;
  margin: 0 0 12px;
}

.original-link {
  display: inline-block;
  font-size: 14px;
  color: #1a73e8;
  margin-bottom: 24px;
}

.content-section {
  margin-bottom: 32px;
}

.content-section h2 {
  font-size: 16px;
  color: #333;
  border-bottom: 2px solid #e0e0e0;
  padding-bottom: 8px;
  margin: 0 0 16px;
}

.content-body p {
  font-size: 15px;
  line-height: 1.8;
  margin: 0 0 12px;
  color: #333;
}

.translation-section {
  background: #fafafa;
  border-radius: 8px;
  padding: 16px;
}

.translation-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.translation-header h2 {
  margin-bottom: 0;
  border-bottom: none;
  padding-bottom: 0;
}

.translate-btn {
  padding: 8px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.translate-btn.idle {
  background: #1a73e8;
  color: #fff;
}

.translate-btn.idle:hover {
  background: #1557b0;
}

.translate-btn.loading {
  background: #e0e0e0;
  color: #666;
  cursor: wait;
}

.translate-btn.success {
  background: #4caf50;
  color: #fff;
  cursor: default;
}

.translate-btn.error {
  background: #d32f2f;
  color: #fff;
}

.translate-btn.error:hover {
  background: #b71c1c;
}

.translate-error {
  color: #d32f2f;
  font-size: 13px;
  margin-top: 8px;
}

.translation-body {
  margin-top: 16px;
}

.translation-placeholder {
  color: #999;
  font-size: 14px;
  margin-top: 16px;
}
</style>
