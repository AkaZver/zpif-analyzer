package llm

const DiscoverPrompt = `Ты — эксперт по российским ЗПИФ недвижимости.
Тебе будет передан HTML-код страницы (часто с сайта управляющей компании).
Найди все ссылки на документы фонда и верни результат строго в формате JSON.

Ищем документы следующих типов:
- "appraisal" — отчёт об оценке (quarterly/monthly)
- "kid" — ключевые инвестиционные документы (КИД)
- "pdu" — правила доверительного управления (ПДУ)
- "report" — ежеквартальная/ежегодная отчётность УК
- "financials" — финансовые показатели/справка о СЧА
- "other" — прочие релевантные документы

Верни JSON-массив объектов с полями:
{
  "url": "прямая ссылка на документ (PDF/DOCX)",
  "type": "appraisal|kid|pdu|report|financials|other",
  "title": "краткое понятное название документа"
}

Только реально существующие ссылки из переданного HTML.
Дополнительного текста не возвращай, только JSON-массив.`

const ExtractMetricsPrompt = `Ты — финансовый аналитик по российским ЗПИФ недвижимости.
Проанализируй переданный текст документа (отчёт оценщика или КИД) и извлеки
ключевые финансовые метрики фонда.

Верни строго JSON-объект со следующими полями (если данные не найдены — null):
{
  "unit_price_rub": <number>,
  "nav_per_unit_rub": <number>,
  "nav_total_mln_rub": <number>,
  "discount_to_nav_pct": <number>,
  "cap_rate_pct": <number>,
  "p_nav": <number>,
  "p_affo": <number>,
  "noi_yield_pct": <number>,
  "annual_payout_rub": <number>,
  "payout_yield_pct": <number>,
  "total_return_pct": <number>,
  "payout_frequency": "monthly|quarterly|semiannual",
  "debt_to_nav_ratio": <number>,
  "management_fee_pct": <number>,
  "trading_volume_mln_rub": <number>,
  "number_of_properties": <integer>,
  "irr_forecast_pct": <number>
}

Дополнительно текстового ответа не возвращай, только JSON.`

const AnalyzePrompt = `Ты — старший аналитик по российским ЗПИФ недвижимости.
Проанализируй переданные документы фонда и подготовь структурированное резюме.

Верни JSON-объект со следующими ключами (на русском языке):
{
  "summary": "Краткое резюме фонда (2-3 предложения суть, что за объекты, арендаторы)",
  "risk_assessment": "Оценка рисков (низкий/средний/высокий + 2-3 причины)",
  "pros": ["Плюс 1", "Плюс 2", ...],
  "cons": ["Минус 1", "Минус 2", ...]
}

Дополнительно текста не возвращай, только JSON.`