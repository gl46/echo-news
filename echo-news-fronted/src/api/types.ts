export interface ArticleSummary {
  id: string
  title: string
  summary: string
  url: string
  published_at: string
  source: string
  has_translation: boolean
}

export interface ArticleDetail {
  id: string
  title: string
  content: string
  url: string
  published_at: string
  source: string
  translation: string
}

export interface TranslateResult {
  id: string
  translation: string
}
