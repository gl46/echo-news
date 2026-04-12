import type { ArticleDetail, ArticleSummary, TranslateResult } from './types'
import {
  fetchArticle as fetchMockArticle,
  translateArticle as translateMockArticle,
} from './mock'

const articleCache = new Map<string, ArticleSummary>()
const translationCache = new Map<string, string>()

async function requestJSON<T>(path: string): Promise<T> {
  const response = await fetch(path, {
    headers: {
      Accept: 'application/json',
    },
  })

  if (!response.ok) {
    throw new Error(`Request failed: ${response.status}`)
  }

  return response.json() as Promise<T>
}

function cacheArticles(items: ArticleSummary[]) {
  for (const item of items) {
    articleCache.set(item.id, item)
  }
}

function buildFallbackDetail(article: ArticleSummary): ArticleDetail {
  const content = article.summary
    ? `${article.summary}\n\nFull article content is not available in the backend yet. Open the original BBC link to read the complete story.`
    : 'Full article content is not available in the backend yet. Open the original BBC link to read the complete story.'

  return {
    id: article.id,
    title: article.title,
    content,
    url: article.url,
    published_at: article.published_at,
    source: article.source,
    translation: translationCache.get(article.id) ?? '',
  }
}

export async function fetchArticles(): Promise<ArticleSummary[]> {
  const articles = await requestJSON<ArticleSummary[]>('/api/articles')
  cacheArticles(articles)
  return articles
}

export async function fetchArticle(id: string): Promise<ArticleDetail> {
  const cachedArticle = articleCache.get(id)
  if (cachedArticle) {
    return buildFallbackDetail(cachedArticle)
  }

  try {
    const articles = await fetchArticles()
    const article = articles.find((item) => item.id === id)
    if (article) {
      return buildFallbackDetail(article)
    }
  } catch {
    // Fall back to the local mock article below if the backend is not reachable.
  }

  return fetchMockArticle(id)
}

export async function translateArticle(id: string): Promise<TranslateResult> {
  const cachedTranslation = translationCache.get(id)
  if (cachedTranslation) {
    return { id, translation: cachedTranslation }
  }

  try {
    const result = await translateMockArticle(id)
    translationCache.set(id, result.translation)
    return result
  } catch {
    const article = articleCache.get(id)
    if (!article) {
      throw new Error('Translation API is not ready yet')
    }

    const placeholderTranslation =
      `翻译接口还没接上，这里先展示一份前端占位内容。\n\n` +
      `标题：${article.title}\n\n` +
      `摘要：${article.summary}`

    translationCache.set(id, placeholderTranslation)
    return { id, translation: placeholderTranslation }
  }
}
