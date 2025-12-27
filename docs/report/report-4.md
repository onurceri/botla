PHASE 4 — Performance & Scalability
1) Ardışık pipeline darboğazı

Risk: Fetch → parse → chunk → embed sıralı.

Etki: Kaynak sayısı arttıkça eğitim süresi lineer artar.

Öneri: Paralel chunk embedding, batch API’ler.

Varsayım: Orta–yüksek veri hacmi.

2) Vector store erişim paterni

Risk: Chat sırasında fazla context çekimi.

Etki: Latency artışı.

Öneri: Top-k sınırlama, cache.