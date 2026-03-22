import type { ArticleSummary, ArticleDetail, TranslateResult } from './types'

const mockArticles: ArticleDetail[] = [
  {
    id: 'bbc-001',
    title: 'Global leaders meet to discuss climate change targets',
    content:
      'World leaders gathered in Geneva on Monday to discuss new climate change targets for the next decade. The summit, which brings together representatives from over 150 countries, aims to establish binding commitments to reduce carbon emissions by 50% before 2035.\n\n"We are at a critical juncture," said the UN Secretary-General in his opening remarks. "The science is clear, and the time for action is now."\n\nThe conference will run for three days, with working groups focusing on renewable energy adoption, deforestation, and industrial emissions. Several developing nations have called for increased financial support to help them transition to cleaner energy sources.\n\nEnvironmental groups have staged peaceful protests outside the venue, urging leaders to go beyond previous pledges that they say have fallen short of what is needed.',
    url: 'https://www.bbc.com/news/world-001',
    published_at: '2026-03-22T10:00:00Z',
    source: 'BBC',
    translation: '',
  },
  {
    id: 'bbc-002',
    title: 'New study reveals high levels of microplastics in drinking water',
    content:
      'A groundbreaking study published in the journal Nature has found that microplastic contamination in drinking water is far more widespread than previously thought. Researchers from the University of Cambridge tested samples from 30 countries and found detectable levels of microplastics in 94% of tap water samples.\n\nThe tiny plastic particles, some smaller than a human red blood cell, are believed to come from a variety of sources including packaging, synthetic clothing fibres, and industrial waste.\n\nWhile the health effects of ingesting microplastics are still being studied, early research suggests potential links to inflammation and hormonal disruption. The World Health Organization has called for urgent further investigation.\n\n"This should be a wake-up call," said Dr. Sarah Chen, the study\'s lead author. "We need to fundamentally rethink how we produce and dispose of plastic."',
    url: 'https://www.bbc.com/news/health-002',
    published_at: '2026-03-22T08:30:00Z',
    source: 'BBC',
    translation: '',
  },
  {
    id: 'bbc-003',
    title: 'Tech giants announce joint AI safety initiative',
    content:
      'Five of the world\'s largest technology companies have announced a joint initiative to develop safety standards for artificial intelligence. The coalition, which includes leading firms from the US, Europe, and Asia, will focus on creating shared testing frameworks and transparency guidelines.\n\nThe announcement comes amid growing public concern about the rapid deployment of AI systems in healthcare, finance, and law enforcement. Critics have pointed to several high-profile cases where AI systems have produced biased or inaccurate results.\n\n"AI has enormous potential to benefit society, but only if we get the safety foundations right," said the initiative\'s newly appointed director. The group plans to release its first set of recommendations within six months.',
    url: 'https://www.bbc.com/news/technology-003',
    published_at: '2026-03-21T15:00:00Z',
    source: 'BBC',
    translation: '',
  },
  {
    id: 'bbc-004',
    title: 'Historic peace agreement signed in East Africa',
    content:
      'After two years of negotiations, rival factions in the Horn of Africa have signed a comprehensive peace agreement that diplomats are calling a historic breakthrough. The deal, brokered with the support of the African Union and the United Nations, includes provisions for power-sharing, disarmament, and the return of displaced populations.\n\nThe signing ceremony took place in Addis Ababa, with representatives from neighbouring countries serving as witnesses. International leaders praised the agreement as a model for conflict resolution on the continent.\n\nHumanitarian organisations have cautioned that the real work lies ahead, noting that previous ceasefires have collapsed. The UN has pledged to deploy a monitoring mission to help ensure compliance with the terms of the agreement.',
    url: 'https://www.bbc.com/news/world-africa-004',
    published_at: '2026-03-21T12:00:00Z',
    source: 'BBC',
    translation: '',
  },
  {
    id: 'bbc-005',
    title: 'Record-breaking heatwave sweeps across Southern Europe',
    content:
      'Southern Europe is experiencing its earliest recorded heatwave, with temperatures exceeding 40°C in parts of Spain, Italy, and Greece. Meteorologists say the event is unprecedented for March and is consistent with climate change projections.\n\nAuthorities in several countries have issued health warnings, particularly for elderly residents and outdoor workers. Schools in parts of Andalusia have shortened their hours, and water restrictions have been imposed in some Greek islands.\n\nThe heatwave has also raised concerns about this year\'s wildfire season. Fire services across the region have been placed on high alert, with additional aircraft and ground crews positioned in vulnerable areas.\n\n"We are seeing summer conditions in spring," said a spokesperson for the European Environment Agency. "This is no longer an anomaly — it is becoming the pattern."',
    url: 'https://www.bbc.com/news/world-europe-005',
    published_at: '2026-03-20T09:00:00Z',
    source: 'BBC',
    translation: '',
  },
]

const mockTranslations: Record<string, string> = {
  'bbc-001':
    '周一，世界各国领导人齐聚日内瓦，讨论未来十年新的气候变化目标。此次峰会汇集了来自150多个国家的代表，旨在建立具有约束力的承诺，在2035年前将碳排放减少50%。\n\n联合国秘书长在开幕致辞中表示："我们正处于关键时刻。科学已经很清楚，现在是采取行动的时候了。"\n\n会议将持续三天，工作组将重点讨论可再生能源应用、森林砍伐和工业排放等议题。多个发展中国家呼吁增加财政支持，帮助其向清洁能源转型。\n\n环保团体在会场外举行了和平抗议，敦促领导人超越此前被认为不够充分的承诺。',
  'bbc-002':
    '发表在《自然》杂志上的一项突破性研究发现，饮用水中的微塑料污染远比之前认为的更加普遍。剑桥大学的研究人员测试了来自30个国家的样本，发现94%的自来水样本中都检测到了微塑料。\n\n这些微小的塑料颗粒，有些比人类红细胞还小，据信来自多种来源，包括包装、合成服装纤维和工业废物。\n\n虽然摄入微塑料对健康的影响仍在研究中，但早期研究表明其可能与炎症和激素干扰有关。世界卫生组织呼吁进行紧急的进一步调查。\n\n该研究的第一作者陈博士说："这应该是一个警钟。我们需要从根本上重新思考我们生产和处理塑料的方式。"',
  'bbc-003':
    '全球五家最大的科技公司宣布了一项联合倡议，旨在为人工智能制定安全标准。该联盟包括来自美国、欧洲和亚洲的领先企业，将重点开发共享测试框架和透明度准则。\n\n该公告发布之际，公众对人工智能系统在医疗、金融和执法领域的快速部署日益担忧。批评人士指出了几起人工智能系统产生偏见或不准确结果的高调案例。\n\n该倡议新任命的负责人表示："人工智能具有造福社会的巨大潜力，但前提是我们必须打好安全基础。"该小组计划在六个月内发布第一批建议。',
}

// Simulated translations storage for the session
const translationCache = new Map<string, string>()

function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

export async function fetchArticles(): Promise<ArticleSummary[]> {
  await delay(300)
  return mockArticles.map((a) => ({
    id: a.id,
    title: a.title,
    summary: a.content.slice(0, 120) + '...',
    url: a.url,
    published_at: a.published_at,
    source: a.source,
    has_translation: translationCache.has(a.id),
  }))
}

export async function fetchArticle(id: string): Promise<ArticleDetail> {
  await delay(200)
  const article = mockArticles.find((a) => a.id === id)
  if (!article) throw new Error('Article not found')
  return {
    ...article,
    translation: translationCache.get(id) ?? '',
  }
}

export async function translateArticle(id: string): Promise<TranslateResult> {
  await delay(1500)
  const translation = mockTranslations[id]
  if (!translation) throw new Error('Translation service unavailable')
  translationCache.set(id, translation)
  return { id, translation }
}
