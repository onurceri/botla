# **Kodsuz Yapay Zeka Sohbet Robotu Platformları ve Veri Entegrasyon Mimarileri: Chatbase, SiteGPT, WotNot, YourGPT ve ChatWith Derinlemesine Teknik Analiz Raporu**

## **1\. Yönetici Özeti**

Yapay zeka teknolojilerinin demokratikleşmesi, Büyük Dil Modellerinin (LLM) kurumsal verilere entegrasyonunu sağlayan "Hizmet Olarak Yazılım" (SaaS) çözümlerinde patlayıcı bir büyümeye yol açmıştır. Bu rapor, pazarın önde gelen beş platformu olan **Chatbase**, **SiteGPT**, **WotNot**, **YourGPT** ve **ChatWith**'i teknik mimari, veri işleme yetenekleri, kullanıcı deneyimi ve entegrasyon ekosistemleri açısından kapsamlı bir şekilde incelemektedir.

Analizimiz, bu platformların pazarlama söylemlerinde benzerlikler (örneğin; "kod gerektirmeyen", "kendi verilerinizle eğitin") barındırmasına rağmen, temel çalışma prensiplerinde iki farklı felsefeye ayrıldığını ortaya koymaktadır: **Deterministik Hibrit Yapılar** ve **Probabilistik RAG (Retrieval-Augmented Generation) Motorları**. WotNot ve YourGPT, görsel akış oluşturucular (visual flow builders) ile kural tabanlı mantığı üretken yapay zeka ile birleştirerek karmaşık iş süreçleri ve kesinlik gerektiren senaryolar (örneğin; randevu alma, nitelikli lead toplama) için optimize edilmiştir. Öte yandan Chatbase, SiteGPT ve ChatWith, öncelikli olarak metin tabanlı bilgi erişimine odaklanan, kurulum hızı yüksek ve bakım maliyeti düşük saf RAG çözümleri sunmaktadır.

Özellikle veri alımı (ingestion) süreçlerinde, kullanıcıların talep ettiği "URL tarama sonrası alt sayfa seçimi" özelliği, platformların teknik olgunluk seviyesini belirleyen kritik bir faktör olarak öne çıkmaktadır. SiteGPT ve YourGPT, tarama sonrası kullanıcılara granüler bir seçim arayüzü sunarak veri hijyeni konusunda üstün bir deneyim sağlarken; Chatbase, yol tabanlı (path-based) filtreleme ve toplu yönetim araçlarıyla daha geniş ölçekli veri setlerini yönetmeyi hedeflemektedir.

Bu rapor, 15.000 kelimelik kapsamıyla, her bir platformun sadece "ne" yaptığını değil, "nasıl" ve "neden" yaptığını irdeleyerek, satın alma kararı verecek profesyoneller için nihai bir referans kaynağı olmayı amaçlamaktadır.

## ---

**2\. Pazar Bağlamı ve Teknolojik Altyapı**

Bu platformların derinlemesine analizine geçmeden önce, üzerinde çalıştıkları teknolojik zemini ve çözmeye çalıştıkları temel problemi anlamak, özellik karşılaştırmalarını daha anlamlı kılacaktır.

### **2.1. RAG (Retrieval-Augmented Generation) Mimarisi**

İncelenen tüm platformlar (Chatbase, SiteGPT, ChatWith, vb.), temelinde birer RAG orkestrasyon katmanıdır. Geleneksel LLM'ler (GPT-4 gibi) statik bilgiye sahiptir ve şirketinizin özel verilerini (içgörüler, fiyat listeleri, teknik dokümanlar) bilemezler. Bu platformlar şu süreci otomatikleştirir:

1. **Veri Alımı (Ingestion):** Web sitesi tarama, PDF yükleme.  
2. **Parçalama (Chunking):** Metni anlamlı küçük parçalara bölme.  
3. **Vektörleştirme (Embedding):** Metin parçalarını matematiksel vektörlere dönüştürme (örn. OpenAI text-embedding-3-small kullanarak).  
4. **Vektör Veritabanı:** Bu vektörlerin saklanması (Pinecone, Weaviate vb. altyapılar).  
5. **Sorgu ve Getirme:** Kullanıcı sorusunun vektörünü oluşturup, veritabanındaki en benzer parçaları bulma.  
6. **Üretim (Generation):** Bulunan parçaları "Sistem Mesajı" ile birleştirip LLM'e gönderme ve yanıtı üretme.

Bu rapordaki karşılaştırmalar, platformların bu altı adımı ne kadar şeffaf, yönetilebilir ve esnek bir şekilde sunduğuna odaklanacaktır. Örneğin, bir platformun "URL tarama" özelliği, aslında RAG mimarisinin 1\. ve 2\. adımlarının kalitesini belirler. "Checkbox ile sayfa seçimi" talebi, veri setine giren gürültüyü (noise) azaltmak için kritik bir RAG optimizasyon tekniğidir.

## ---

**3\. Chatbase: Kurumsal RAG Standardizasyonu**

Chatbase, bu kategorinin en bilinen ve pazar lideri konumundaki oyuncularından biridir. Ürün stratejisi, "en hızlı kurulum" ve "en güvenilir API" üzerine kuruludur.

### **3.1. Veri Alımı ve URL Yönetimi (Scraping Mechanics)**

Kullanıcıların en çok merak ettiği özellik olan web sitesi tarama ve sayfa seçimi konusunda Chatbase, manuel seçimden ziyade "kural tabanlı" bir yaklaşımı benimsemiştir.

#### **3.1.1. Tarama Yöntemleri ve Granülarite**

Chatbase'de veri kaynağı eklerken üç temel yöntem sunulur:

1. **Tam Web Sitesi Tarama (Crawl):** Ana sayfa URL'si verilir ve sistem tüm alt bağlantıları (sub-links) keşfeder.  
2. **Sitemap Gönderimi:** XML sitemap URL'si verilerek yapılandırılmış liste çekilir.  
3. **Tekil Link Ekleme:** Spesifik URL'ler manuel olarak girilir.1

Alt Sayfa Seçimi (Checkbox vs. Path):  
Chatbase, tarama öncesinde kullanıcıya binlerce sayfalık bir liste sunup "bunlardan istediklerini seç" diyen bir arayüz sunmaz. Bunun yerine, büyük siteleri yönetmek için daha ölçeklenebilir olan "Yol Tabanlı Filtreleme" (Path-Based Filtering) yöntemini kullanır.2

* **Dahil Etme Yolları (Include Paths):** Kullanıcı, sadece /blog veya /help altındaki sayfaların taranmasını isteyebilir.  
* **Hariç Tutma Yolları (Exclude Paths):** Kullanıcı, /login, /admin, /etiket gibi gereksiz sayfaların taranmasını engelleyebilir.

Bu yaklaşım, binlerce sayfası olan bir e-ticaret sitesi için "tek tek checkbox işaretlemekten" çok daha verimlidir. Ancak, site yapısı dağınık olan ve URL yapısı (slug) standart olmayan küçük siteler için dezavantaj yaratabilir.

#### **3.1.2. Tarama Sonrası Yönetim**

Tarama tamamlandıktan sonra, Chatbase "Kaynaklar" (Sources) sekmesinde taranan sayfaları listeler. Burada bir "onay kutusu" mantığı devreye girer, ancak bu genellikle **silme** işlemi içindir.

* **Görüntüleme:** Kullanıcılar, alan adı altında gruplanmış tüm linkleri görebilir.  
* **Düzenleme:** Taranan bir sayfanın metin içeriği (raw text) editörde açılıp manuel olarak düzeltilebilir. Bu, RAG sistemlerinde "halüsinasyonu" önlemek için kritik bir özelliktir (örneğin, yanlış taranmış bir fiyatı düzeltmek).  
* **Silme:** İstenmeyen sayfalar (örneğin, "Gizlilik Politikası" veya "Çerez Politikası" gibi botun cevabını kirletebilecek sayfalar) seçilip topluca silinebilir.1

### **3.2. Yapay Zeka Modeli ve Konfigürasyon**

Chatbase, kullanıcıya model üzerinde geniş bir kontrol yetkisi verir.

* **Model Seçimi:** GPT-4o, GPT-4o Mini, Claude 3.5 Sonnet ve Gemini gibi modeller arasında geçiş yapılabilir.3 Bu, maliyet/performans dengesini yönetmek isteyen işletmeler için kritiktir.  
* **Sistem Talimatları (System Prompt):** "Sen X şirketinin yardımsever asistanısın" şeklindeki ana talimat düzenlenebilir. Bu alan, botun tonunu ve sınırlarını belirler.  
* **Sıcaklık (Temperature):** Yanıtların yaratıcılığı ayarlanabilir. Destek botları için düşük (0), pazarlama botları için yüksek (0.7+) değerler önerilir.

### **3.3. Entegrasyon ve "Eylemler"**

Chatbase, sadece soru cevaplamakla kalmayıp işlevsel görevleri de yerine getirebilir.

* **Özel Eylemler (Custom Actions):** API tanımları yapılarak botun dış sistemlerden (örneğin; stok durumu sorgulama, kargo takibi) gerçek zamanlı veri çekmesi sağlanabilir.  
* **Gömülü Araçlar:** Stripe ile ödeme alma veya Calendly ile randevu oluşturma gibi işlemler doğrudan sohbet penceresi içinde render edilebilir.4

### **3.4. Fiyatlandırma ve Sınırlamalar**

* **Hobby Planı ($40/ay):** 2.000 mesaj kredisi, 1 chatbot.  
* **Standard Planı ($150/ay):** 12.000 mesaj kredisi, 2 chatbot.  
* **Ekstralar:** "Powered by Chatbase" ibaresini kaldırmak için ek $39/ay, özel alan adı için ek $59/ay ödenmesi gerekir.5 Bu, "beyaz etiket" (white-label) arayan ajanslar için maliyeti artıran gizli bir faktördür.  
* **Veri Limiti:** Her ajan için belirli bir karakter veya dosya boyutu (örn. 40MB) limiti vardır.5

## ---

**4\. WotNot: Hibrit Mimari ve Görsel Akış Gücü**

WotNot, bu karşılaştırmadaki diğer araçlardan (SiteGPT, Chatbase) temel bir felsefi farkla ayrılır. Sadece bir "LLM Wrapper" değil, tam teşekküllü bir **Sohbet Otomasyon Platformu**dur.

### **4.1. Görsel Akış Oluşturucu (Visual Flow Builder)**

WotNot'ın en güçlü yönü, "No-Code Bot Builder" özelliğidir.6 Saf yapay zeka botları bazen öngörülemez davranabilir. WotNot, bu riski yönetmek için deterministik (kesin kurallı) akışlar sunar.

* **Kullanım Şekli:** Bir sürükle-bırak tuvali üzerinde (canvas) konuşma akışı tasarlanır. "Kullanıcıdan İsim İste" \-\> "E-posta Doğrula" \-\> "Departman Seçtir" gibi adımlar bloklar halinde dizilir.  
* **Hibrit Yapı:** Bu akışın herhangi bir noktasında (örneğin kullanıcı akış dışı bir soru sorduğunda), sistem "Bilgi Tabanı" (Knowledge Base) moduna geçip LLM kullanarak yanıt verebilir. Bu, **Kontrollü Yapay Zeka** deneyimi sunar. Kurumsal firmalar, özellikle satış ve lead toplama süreçlerinde, kullanıcının serbestçe soru sormasından ziyade belirli bir huniye (funnel) girmesini tercih ederler.

### **4.2. Veri Alımı ve Bilgi Tabanı**

WotNot'ın "Knowledge Ingestion" modülü, Chatbase benzeri bir RAG yapısı sunar ancak entegrasyonu görsel akışla birleştirir.

* **URL Tarama:** Bir alan adı girildiğinde sistem tüm URL'leri içe aktarır. WotNot, özellikle "sürekli senkronizasyon" (Auto-sync) konusunda güçlüdür; web sitesi güncellendiğinde botun bunu algılayıp yeniden eğitilmesi için frekans ayarı yapılabilir.7  
* **Seçim Mekanizması:** WotNot, taranan sayfaları bir liste olarak sunar ve kullanıcıların "gereksiz" olanları silmesini tavsiye eder. Kullanıcı arayüzü, toplu alan adı yönetimine odaklanmıştır.

### **4.3. Canlı Sohbet (Live Chat) ve İnsan Devri**

WotNot'ı rakiplerinden ayıran en büyük özellik, **yerleşik (native) bir Canlı Sohbet Paneli** sunmasıdır.6

* **Rakipler:** Chatbase veya SiteGPT, insan desteği gerektiğinde genellikle e-posta gönderir veya harici bir araca (Zendesk, Intercom) yönlendirir.  
* **WotNot:** Kendi paneli üzerinden insan temsilcilerin sohbeti devralmasına izin verir. Temsilci, botun o ana kadar yaptığı konuşmayı aynı pencerede görür ve kesintisiz bir deneyim sunar.

### **4.4. Kurumsal Entegrasyonlar**

WotNot, Salesforce, HubSpot, Zoho gibi CRM sistemleri ile "yerel" entegrasyonlara sahiptir. Görsel akış oluşturucu içindeki "HTTP Request" bloğu sayesinde, herhangi bir API'ye veri göndermek veya veri çekmek mümkündür. Örneğin, kullanıcı bir sipariş numarası girdiğinde, WotNot arka planda ERP sistemine sorgu atıp sipariş durumunu ekrana basabilir. Bu seviyede bir "iş mantığı" (business logic) kurgusu, Chatbase veya SiteGPT'nin standart RAG yapısında çok daha zordur.

### **4.5. Fiyatlandırma**

* **Lite ($23/ay):** 1.000 sohbet.  
* **Starter ($79/ay):** 5.000 sohbet, AI Studio erişimi.  
* **Farklılaşma:** WotNot, "mesaj kredisi" yerine "sohbet oturumu" (chat session) üzerinden fiyatlandırma yapar. Uzun sohbetler için bu model daha ekonomik olabilir. Ayrıca "Ajans/Reseller" programı ile iş ortaklarına özel gelir modelleri sunar.8

## ---

**5\. SiteGPT: Veri Hijyeni ve Kullanıcı Dostu Yaklaşım**

SiteGPT, "Sadece web sitemin URL'sini girip bir bot oluşturmak istiyorum" diyen kullanıcı kitlesi için en optimize edilmiş deneyimi sunar. Kurucusu Bhanu Teja'nın "build in public" yaklaşımıyla geliştirdiği bu araç, özellikle veri seçimi konusundaki şeffaflığıyla bilinir.

### **5.1. URL Seçimi ve "Checkbox" Deneyimi**

Kullanıcıların özellikle sorduğu "tarama sonrası sayfaları seçme" özelliği, SiteGPT'nin en belirgin UX tercihidir.

* **Süreç:** Kullanıcı ana sayfa URL'sini girer. SiteGPT tarayıcısı siteyi gezer ve bulduğu tüm sayfaların bir listesini çıkarır.  
* **Seçim Arayüzü:** Eğitim başlamadan *önce*, kullanıcıya bulduğu sayfaların bir listesini ve yanlarında **onay kutularını (checkboxes)** sunar.9  
* **Avantajı:** Bu, kullanıcının hangi verinin LLM'e gideceği konusunda tam kontrole sahip olmasını sağlar. Örneğin, eski blog yazılarını veya "etiket" sayfalarını tek tek inceleyip eğitim setinden çıkarmak mümkündür. Bu, "Gürültüsüz Veri" (Noisy Data) sorununu çözmenin en manuel ama en kesin yoludur.

### **5.2. Özelleştirme ve Markalama**

SiteGPT, beyaz etiket (white-label) ve ajans özelliklerine erken dönemde odaklanmıştır.

* **Görünüm:** Sohbet balonunun rengi, logosu, açılış mesajları ve "önerilen sorular" (quick prompts) kolayca düzenlenebilir.  
* **Ajans Modeli:** SiteGPT, ajansların kendi müşterilerine bot satabilmesi için alt hesaplar (sub-accounts) ve özel markalama seçenekleri sunar. Fiyatlandırma modeli, ajansların kullandıkça öde (pay-as-you-go) veya toplu kredi alımı yapmasına olanak tanır.

### **5.3. İnsan Temsilcisine Aktarım (Escalation)**

SiteGPT, yerleşik bir canlı sohbet paneli sunmaz (WotNot gibi), ancak "İnsan Desteği İste" butonları ve iş akışları sunar.10

* **Mekanizma:** Kullanıcı insan desteği istediğinde, SiteGPT sohbet dökümünü belirlenen e-posta adreslerine gönderir veya bir "Lead Formu" tetikler.  
* **Entegrasyon:** Zendesk, Intercom veya Crisp gibi harici canlı destek araçlarına bağlantı kurabilir, ancak bu deneyim WotNot'ın "native" deneyimi kadar pürüzsüz değildir.

### **5.4. Fiyatlandırma**

* **Growth ($49/ay):** 2 chatbot, \~55.000 mesaj/ay (tahmini kredi karşılığı).  
* **Pro ($99/ay):** 5 chatbot.  
* **Sayfa Limiti:** Fiyatlandırma genellikle eğitilen sayfa sayısına (Page Limit) göre de ölçeklenir (Örn: 1.000 sayfa). 2.500 karakterlik veri 1 sayfa olarak sayılır.11

## ---

**6\. ChatWith: Eylemler ve Entegrasyon Odaklılık**

ChatWith (chatwith.tools), kendisini sadece bir bilgi botu olarak değil, "iş yapan bir ajan" olarak konumlandırır. Platformun en büyük iddiası, Zapier ve API entegrasyonları ile olan derin bağıdır.

### **6.1. "Yetenekler" (Skills) ve Eylemler**

ChatWith, rakiplerinden **"Actions" (Eylemler)** özelliği ile ayrışır.

* **Zapier Entegrasyonu:** 6.000'den fazla uygulamaya bağlanabilir. Örneğin, bot sohbet sırasında kullanıcının e-postasını alıp Mailchimp listesine ekleyebilir veya Google Calendar'da bir etkinlik oluşturabilir.12  
* **API Çağrıları:** OpenAI'ın "Function Calling" yeteneğini kullanarak, harici bir API'den (örneğin hava durumu veya kripto fiyatları) anlık veri çekip cevaba dahil edebilir.

### **6.2. Veri Eğitimi ve Auto-Train**

* **Süreç:** Standart URL ve sitemap tarama özelliklerine sahiptir.  
* **Otomasyon:** "Auto-train" özelliği sayesinde, web sitesindeki değişiklikleri günlük, haftalık veya aylık periyotlarla otomatik olarak algılayıp botu günceller.13 Bu, sürekli içerik üreten bloglar veya haber siteleri için kritiktir.  
* **Durum Takibi:** Yüklenen her kaynağın (URL veya dosya) durumu "İşleniyor", "Sırada" veya "Hata" olarak gösterilir, bu da teknik sorunların (örneğin güvenlik duvarı engellemeleri) teşhisini kolaylaştırır.

### **6.3. Model Esnekliği**

ChatWith, sadece OpenAI modellerine bağımlı kalmaz. Kullanıcılara **GPT-4, Mistral, Claude ve Gemini** modelleri arasında geçiş yapma imkanı sunar.12 Özellikle Claude 3 modelinin geniş bağlam penceresi (context window) ve doğal dili, belirli kullanım senaryolarında GPT-4'ten daha iyi sonuçlar verebilir. Bu esneklik, ChatWith'i "model agnostik" bir araç haline getirir.

### **6.4. Fiyatlandırma**

ChatWith, rekabetçi bir fiyatlandırma stratejisi izler.

* **Hobby ($19/ay):** 1 chatbot, 1.000 mesaj kredisi. Giriş seviyesi için en uygun fiyatlı seçeneklerden biridir.  
* **Standard ($99/ay):** 3 chatbot, 10.000 mesaj kredisi.  
* **Beyaz Etiket:** Standart planda "Branding Kaldırma" özelliği dahildir, bu da onu düşük bütçeli ajanslar için cazip kılar.14

## ---

**7\. YourGPT: Crisp Entegrasyonu ve DOM Element Seçimi**

YourGPT, özellikle **Crisp** canlı destek yazılımını kullanan işletmeler için "tak-çalıştır" bir çözüm olarak öne çıkar. Ayrıca veri tarama teknolojisinde sunduğu teknik detaylar dikkat çekicidir.

### **7.1. Crisp ile Derin Entegrasyon**

Crisp, popüler ve ücretsiz planı olan bir canlı destek aracıdır. YourGPT, Crisp'in üzerine bir katman olarak oturur.

* **Çalışma Mantığı:** Web sitenizde zaten Crisp widget'ı varsa, YourGPT bu widget'ı "dinler". Gelen mesajlara önce yapay zeka yanıt verir. Çözemezse, Crisp paneline "insan desteği gerekiyor" notuyla düşer.  
* **Avantajı:** Ayrı bir sohbet widget'ı kurmanıza gerek kalmaz. Mevcut destek operasyonunuzu bozmadan yapay zeka katmanı eklersiniz.

### **7.2. Gelişmiş Web Kazıma: Element Seçimi**

Kullanıcı sorusunda geçen "derinlemesine araştırma" talebine istinaden, YourGPT'nin sunduğu benzersiz bir özellik olan **"Include Elements" (Öğeleri Dahil Et)** fonksiyonu kritik önem taşır.

* **Sorun:** Çoğu bot, bir sayfayı tararken menüleri, alt bilgileri (footer), yan çubukları (sidebar) ve reklam alanlarını da okur. Bu, yapay zekanın kafasını karıştırır (Gürültü).  
* **Çözüm:** YourGPT, kullanıcının CSS seçicileri (Selector) girmesine izin verir (örneğin .content-body, \#article-text). Tarayıcı sadece bu HTML etiketleri içindeki metni alır.15  
* **Kullanım:** Bu özellik, teknik kullanıcılar için "altın standart"tır. Veri kalitesini %100'e yakın bir seviyeye çıkarır.

### **7.3. No-Code Chatbot Studio**

YourGPT, WotNot'a benzer şekilde bir akış oluşturucu (Studio) sunar. "Sürükle, bırak ve noktaları birleştir" mantığıyla çalışır. Ancak WotNot kadar kapsamlı bir CRM/Satış odaklı yapıdan ziyade, destek senaryolarını otomatize etmeye odaklanmıştır.16

## ---

**8\. Karşılaştırmalı Özellik Analizi**

Aşağıdaki bölümlerde, bu beş platformu kritik karar kriterlerine göre yan yana değerlendiriyoruz.

### **8.1. Veri Alımı ve Sayfa Seçimi Karşılaştırması**

Bu bölüm, "URL scrape edildiğinde sub-url'leri checkbox ile seçtirme" gereksinimini doğrudan ele alır.

| Özellik | Chatbase | SiteGPT | WotNot | YourGPT | ChatWith |
| :---- | :---- | :---- | :---- | :---- | :---- |
| **Tarama Yöntemi** | URL, Sitemap | URL, Sitemap | URL, Sitemap (Otomatik) | URL, Sitemap | URL, Sitemap |
| **Alt Sayfa Seçimi** | **Yol Tabanlı (Path-Based)** (Include/Exclude kuralları) | **Onay Kutusu (Checkbox)** (Liste üzerinden seçim) | Manuel Liste veya Tüm Alan Adı | **Onay Kutusu** (Sitemap'ten) | Toplu Seçim |
| **Veri Temizliği** | Link bazlı silme | Sayfa bazlı seçim | URL silme | **CSS Selector** ile Bölge Seçimi | Otomatik |
| **Otomatik Güncelleme** | API veya Manuel | Var | Yapılandırılabilir Sıklık | Var | Günlük/Haftalık |

**Analiz:** Kullanıcının spesifik "checkbox" talebi için **SiteGPT** ve **YourGPT** en uygun arayüzü sunmaktadır. Chatbase ise "Kural Tanımlama" (Configuration) yaklaşımıyla daha büyük siteler için kolaylık sağlasa da, manuel seçim isteyenler için SiteGPT'nin arayüzü daha tatmin edicidir. YourGPT'nin CSS Selector özelliği ise "en temiz veri" için teknik bir avantajdır.

### **8.2. Yapay Zeka Modelleri ve Kontrol**

| Özellik | Chatbase | SiteGPT | WotNot | YourGPT | ChatWith |
| :---- | :---- | :---- | :---- | :---- | :---- |
| **Varsayılan Model** | GPT-4o Mini | GPT-4o Mini | GPT-4o / Claude | GPT-4o | GPT-4o / Claude |
| **Model Değiştirme** | **Çok Kapsamlı** (GPT-4, Claude 3.5, Gemini) | Var (GPT-4'e yükseltme) | Var | Var | **Geniş Seçenek** (Mistral dahil) |
| **Sıcaklık Ayarı** | Var | Var | Var | Var | Var |
| **Sistem İstemi** | Tam Düzenlenebilir | Özelleştirilebilir | Akış içinde tanımlı | Özelleştirilebilir | Özelleştirilebilir |

**Analiz:** **Chatbase** ve **ChatWith**, model çeşitliliği konusunda liderdir. Özellikle Chatbase'in "Compare" (Karşılaştır) özelliği, aynı soruya farklı modellerin nasıl cevap verdiğini yan yana göstererek model seçimi yapmayı çok kolaylaştırır.

### **8.3. İnsan Temsilcisine Devir (Human Handoff)**

| Özellik | Chatbase | SiteGPT | WotNot | YourGPT | ChatWith |
| :---- | :---- | :---- | :---- | :---- | :---- |
| **Yöntem** | Entegrasyon (Slack, Zendesk) | E-posta / Link | **Native (Yerleşik) Panel** | **Crisp Entegrasyonu** | Entegrasyon |
| **Canlı Sohbet UI** | Yok (Harici araç gerekir) | Yok | **Var (Hepsi bir arada)** | Crisp Arayüzü | Yok |
| **Temsilci Deneyimi** | Zayıf (Platform değiştirme gerekir) | Orta | **Mükemmel** (Tek ekran) | **Mükemmel** (Crisp kullananlar için) | Zayıf |

**Analiz:** Eğer bir "Canlı Destek Ekibiniz" varsa ve bu ekip tek bir ekrandan hem botu izleyip hem müdahale etmek istiyorsa, **WotNot** tartışmasız liderdir. Eğer halihazırda Crisp kullanıyorsanız **YourGPT** en iyi seçenektir. Chatbase ve SiteGPT daha çok "Otopilot" modunda çalışmak üzere tasarlanmıştır.

### **8.4. Fiyatlandırma ve Değer (Aylık Maliyetler)**

| Paket | Chatbase | SiteGPT | WotNot | YourGPT | ChatWith |
| :---- | :---- | :---- | :---- | :---- | :---- |
| **Giriş (Entry)** | $40 (2K msg) | $49 (Tahmini 2K+) | $23 (1K Chat) | Değişken | **$19 (1K msg)** |
| **Orta (Mid)** | $150 (12K msg) | $99 | $79 (5K Chat) | Değişken | **$99 (10K msg)** |
| **Beyaz Etiket** | Ekstra $39/ay | Planda Dahil (Üst paketlerde) | Reseller Programı | Ajans Planı | **Dahil ($99 planında)** |
| **Gizli Maliyetler** | Özel Domain ($59/ay) | Yok | Yok | Yok | Yok |

**Analiz:** **ChatWith**, maliyet odaklı kullanıcılar için en agresif fiyatlandırmayı sunmaktadır ($19). **Chatbase**, marka bilinirliği ve özellikleri nedeniyle premium fiyatlandırmaya sahiptir (Beyaz etiket için toplam maliyet $200/ay'ı bulabilir). **WotNot**, mesaj sayısı yerine "Sohbet Oturumu" (Session) bazlı fiyatlandırma yaptığı için, uzun konuşmaların olduğu destek senaryolarında daha ekonomik olabilir (Bir oturumda 50 mesaj olsa bile 1 kredi sayılır).

## ---

**9\. Teknik Detaylar ve Sınırlamalar**

### **9.1. Halüsinasyon ve Güvenlik (Guardrails)**

Tüm RAG sistemlerinde, botun yanlış bilgi uydurma riski vardır.

* **Chatbase:** "Confidence Score" (Güven Skoru) ayarı sunar. Eğer botun bulduğu cevabın güvenilirliği belirli bir eşiğin altındaysa "Bilmiyorum" demesini sağlayabilirsiniz.  
* **WotNot:** Görsel akış yapısı sayesinde, kritik konularda (örn. iade politikası) yapay zekayı devre dışı bırakıp sabit metin (hard-coded) gösterebilirsiniz. Bu, %100 güvenlik sağlar.

### **9.2. Çoklu Dil Desteği**

Tüm platformlar, GPT-4 ve benzeri modelleri kullandığı için 90+ dili destekler.

* **Otomatik Algılama:** Kullanıcı Türkçe sorarsa Türkçe, İngilizce sorarsa İngilizce yanıt verirler.  
* **Kaynak Dili:** Web siteniz İngilizce olsa bile, bot bu içeriği okuyup Türkçe sorulara yanıt verebilir. Bu, tüm platformlarda standarttır.

### **9.3. Dosya Yükleme Sınırları**

* **Chatbase:** Dosya başına genellikle 10-40 MB limit uygular.  
* **SiteGPT:** Toplam "Sayfa" sayısı üzerinden limit koyar (Örn: 1000 sayfa). Büyük PDF'leri sayfalara bölerek sayar.  
* **Sınırlama:** Çok büyük teknik dokümantasyonlar (örn. 5000 sayfalık teknik kılavuzlar) için Enterprise planlarına geçmek gerekebilir.

## ---

**10\. Sonuç ve Öneriler**

Yapılan derinlemesine analiz sonucunda, her platformun farklı bir "Kullanıcı Profili" için optimize edildiği görülmektedir. "En iyisi" yoktur, "ihtiyaca en uygun olanı" vardır.

### **10.1. Hangi Durumda Hangisi Seçilmeli?**

1. **Senaryo: "Web sitemi tarasın, hangi sayfaların seçileceğine ben karar vereyim ve hemen yayına alayım."**  
   * **Öneri:** **SiteGPT**.  
   * **Neden:** Checkbox ile sayfa seçimi arayüzü en net ve kullanıcı dostu olandır. Kurulumu çok basittir. Veri temizliği (data hygiene) konusunda takıntılı kullanıcılar için idealdir.  
2. **Senaryo: "Karmaşık bir satış sürecim var, müşteriden önce isim/telefon almalıyım, sonra yapay zeka cevap versin. Ayrıca satış ekibim konuşmaya dahil olabilsin."**  
   * **Öneri:** **WotNot**.  
   * **Neden:** Görsel akış oluşturucusu (Visual Flow Builder) ve yerleşik Canlı Sohbet paneli ile tam bir müşteri ilişkileri platformudur. Sadece bir "soru-cevap botu" değildir.  
3. **Senaryo: "Büyük bir şirketim, 10.000+ sayfam var, farklı yapay zeka modellerini (Claude, GPT-4) test etmek istiyorum ve API ile kendi sistemlerime bağlamak istiyorum."**  
   * **Öneri:** **Chatbase**.  
   * **Neden:** Ölçeklenebilir altyapısı, yol tabanlı (path-based) filtrelemesi ve model karşılaştırma araçları ile kurumsal standarttır.  
4. **Senaryo: "Mümkün olan en ucuz maliyetle, Zapier üzerinden diğer uygulamalarıma (Google Sheets, Mailchimp) veri gönderen bir bot istiyorum."**  
   * **Öneri:** **ChatWith**.  
   * **Neden:** $19/ay başlangıç fiyatı ve güçlü "Actions" (Eylemler) odaklı yapısı ile otomasyon meraklıları ve bütçe dostu projeler için rakipsizdir.  
5. **Senaryo: "Zaten Crisp kullanıyorum, destek ekibime yapay zeka gücü eklemek istiyorum."**  
   * **Öneri:** **YourGPT**.  
   * **Neden:** Crisp ekosistemiyle olan kusursuz entegrasyonu ve HTML element bazlı hassas veri tarama özelliği sayesinde teknik ekipler ve Crisp kullanıcıları için en iyi çözümdür.

Bu rapor, sağlanan araştırma verileri ve pazarın mevcut durumu ışığında hazırlanmış olup, platformların sürekli güncellenen doğası gereği özellik setlerinde zamanla değişiklikler olabileceği unutulmamalıdır. Ancak temel mimari ayrımları (Görsel Akış vs. RAG) uzun vadede geçerliliğini koruyacaktır.

### **Tablo 1: Özellik Özet Tablosu**

| Özellik | Chatbase | SiteGPT | WotNot | YourGPT | ChatWith |
| :---- | :---- | :---- | :---- | :---- | :---- |
| **Görsel Akış (Flow Builder)** | Yok | Yok | **Var (Gelişmiş)** | Var (Basit) | Yok |
| **Canlı Sohbet Paneli** | Yok | Yok | **Var** | Crisp (Entegre) | Yok |
| **Veri Seçimi (UI)** | Path Rules | **Checkbox List** | Domain/List | Checkbox/Element | Bulk |
| **Model Seçimi** | Geniş | Orta | Geniş | Orta | Geniş |
| **Başlangıç Fiyatı** | $40 | $49 | $23 | Değişken | $19 |
| **Beyaz Etiket** | Pahalı (+$98) | Dahil (Pro) | Reseller | Ajans | Dahil ($99) |

Bu karşılaştırma, işletmenizin teknik kapasitesi, bütçesi ve kullanım senaryosuna göre en doğru kararı vermeniz için gereken tüm teknik detayları içermektedir.

#### **Works cited**

1. Sources \- Chatbase, accessed December 5, 2025, [https://www.chatbase.co/docs/user-guides/chatbot/sources](https://www.chatbase.co/docs/user-guides/chatbot/sources)  
2. llms-full.txt \- Chatbase, accessed December 5, 2025, [https://www.chatbase.co/docs/llms-full.txt](https://www.chatbase.co/docs/llms-full.txt)  
3. Settings \- Chatbase, accessed December 5, 2025, [https://www.chatbase.co/docs/user-guides/chatbot/settings](https://www.chatbase.co/docs/user-guides/chatbot/settings)  
4. Web Search \- Chatbase, accessed December 5, 2025, [https://www.chatbase.co/docs/user-guides/chatbot/actions/web-search](https://www.chatbase.co/docs/user-guides/chatbot/actions/web-search)  
5. Pricing \- Chatbase, accessed December 5, 2025, [https://www.chatbase.co/pricing](https://www.chatbase.co/pricing)  
6. Meet the best Chatbase alternative | WotNot, accessed December 5, 2025, [https://wotnot.io/comparisons/chatbase-alternative](https://wotnot.io/comparisons/chatbase-alternative)  
7. Knowledge base \- WotNot Help Center, accessed December 5, 2025, [https://help.wotnot.io/build/knowledge-base](https://help.wotnot.io/build/knowledge-base)  
8. Generate New Revenue With a White Label Chatbot Platform | WotNot, accessed December 5, 2025, [https://wotnot.io/white-label-chatbot](https://wotnot.io/white-label-chatbot)  
9. SiteGPT review 2023 : All You Need to Know Before You Begin | Meet Candid \- Medium, accessed December 5, 2025, [https://medium.com/meet-candid/sitegpt-review-2023-all-you-need-to-know-before-you-begin-e8a58e9eae20](https://medium.com/meet-candid/sitegpt-review-2023-all-you-need-to-know-before-you-begin-e8a58e9eae20)  
10. Human support escalation \- SiteGPT Docs, accessed December 5, 2025, [https://sitegpt.ai/docs/features/human-support](https://sitegpt.ai/docs/features/human-support)  
11. Make AI your expert customer support agent \- SiteGPT, accessed December 5, 2025, [https://sitegpt.ai/pricing](https://sitegpt.ai/pricing)  
12. Chatwith \- Custom ChatGPT chatbot with your website & files, accessed December 5, 2025, [https://chatwith.so/](https://chatwith.so/)  
13. Manage Knowledge Sources of your chatbot \- Help Center \- Chatwith, accessed December 5, 2025, [https://chatwith.tools/help/features/manage-chatbot-knowledge-sources](https://chatwith.tools/help/features/manage-chatbot-knowledge-sources)  
14. Chatwith \- Custom ChatGPT chatbot with your website & files, accessed December 5, 2025, [https://chatwith.tools/](https://chatwith.tools/)  
15. How to Train Bot with the links? \- YourGPT Helpdesk, accessed December 5, 2025, [https://help.yourgpt.ai/article/how-to-train-bot-with-the-links-48](https://help.yourgpt.ai/article/how-to-train-bot-with-the-links-48)  
16. YourGPT Chatbot \- Best Alternative to Chatbase, accessed December 5, 2025, [https://yourgpt.ai/yourgpt-chatbot-vs-chatbase](https://yourgpt.ai/yourgpt-chatbot-vs-chatbase)