import type { ArticleDetail, ArticleSummary, TranslateResult } from './types'
import { fetchArticle as fetchMockArticle } from './mock'

const articleCache = new Map<string, ArticleSummary>()
const translationCache = new Map<string, string>()

async function requestJSON<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(path, {
    ...options,
    headers: {
      Accept: 'application/json',
      ...options.headers,
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
  return {
    id: article.id,
    title: article.title,
    content: article.summary || article.title,
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

  const result = await requestJSON<TranslateResult>(`/api/translate/${encodeURIComponent(id)}`, {
    method: 'POST',
  })

  translationCache.set(id, result.translation)
  return result
}
