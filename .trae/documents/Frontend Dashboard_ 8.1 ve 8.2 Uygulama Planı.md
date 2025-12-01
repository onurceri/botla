## Amaç ve Kapsam
- 8.1 React Project Setup ve 8.2 Folder Structure adımlarını uygulayarak `frontend` Vite + React tabanlı bir dashboard kurmak.
- TailwindCSS, Router, Query, Axios ve UI bileşenleri için başlangıç yapılandırmalarını yapmak.
- Dosya hiyerarşisini ve temel iskelet dosyalarını oluşturmak; uygulama derlenebilir ve çalıştırılabilir olacak.

## Ön Koşullar
- Node.js 18+ ve npm kurulu.
- Depo kökünde yeni `frontend/` dizini oluşturulacak.

## Adım 1: Vite React Proje Kurulumu
1. `frontend` dizinini oluştur ve içine geç.
2. Vite kurulumunu başlat: `npm create vite@latest . -- --template react`
3. Bağımlılıklar: `npm install`
4. TypeScript dev bağımlılıkları: `npm install -D typescript @types/react @types/react-dom`
5. Not: İstersek doğrudan TS template (`--template react-ts`) kullanılabilir; planda dokümanla uyum için TS dev bağımlılıkları eklenerek JS → TS dönüşümü yapılacak.

## Adım 2: Tailwind ve UI Kurulumu
1. Tailwind: `npm install tailwindcss postcss autoprefixer` ve `npx tailwindcss init -p`
2. `tailwind.config.js` içinde `content` yollarını `./index.html`, `./src/**/*.{ts,tsx}` şeklinde ayarla.
3. `src/index.css` içine Tailwind direktiflerini ekle: `@tailwind base; @tailwind components; @tailwind utilities;`
4. UI & ikonlar:
   - `lucide-react` kurulumu: `npm install lucide-react`
   - shadcn bileşenleri için CLI yaklaşımı kullanılacak: `npx shadcn@latest init` (gerekirse). Üretilen `components/ui/*` bileşenleri `components/shared/*` altında yeniden ihrac edilecek veya basit Tailwind tabanlı versiyonlar yazılacak.
   - Radix: İhtiyaç duyulan paketler (ör. `@radix-ui/react-dialog`, `@radix-ui/react-dropdown-menu`) sayfa/bileşen gereksinimine göre eklenecek. Dokümandaki `@radix-ui/react-*` wildcard ifadesi yerine spesifik paketler kurulacak.

## Adım 3: Proje Yapısı Oluşturma
`src/` altında dokümanla birebir hiyerarşi oluşturulacak:
- `components/layout/Header.tsx`, `Sidebar.tsx`, `MainLayout.tsx`
- `components/chatbot/ChatbotCard.tsx`, `ChatbotForm.tsx`, `SourceUploader.tsx`
- `components/shared/Button.tsx`, `Card.tsx`, `Modal.tsx`
- `pages/LoginPage.tsx`, `DashboardPage.tsx`, `ChatbotsPage.tsx`, `ChatbotDetailPage.tsx`, `AnalyticsPage.tsx`, `SettingsPage.tsx`
- `api/client.ts`, `auth.ts`, `chatbot.ts`, `source.ts`, `analytics.ts`
- `hooks/useAuth.ts`, `useChatbots.ts`, `usePagination.ts`
- `types/auth.ts`, `chatbot.ts`, `api.ts`, `index.ts`
- `utils/localStorage.ts`, `format.ts`, `validators.ts`
- `App.tsx`, `main.tsx`
Ayrıca kökte `public/`, `vite.config.ts`, `tailwind.config.js` ve `package.json` (Vite’in oluşturduğu) bulunacak.

## Adım 4: Router ve Layout
- `react-router-dom` kurulumu: `npm install react-router-dom`
- Temel router kurgusu `App.tsx` içinde yapılacak.
- `MainLayout` ile `Header` + `Sidebar` birleşecek; `DashboardPage` gibi sayfalar layout içinde render edilecek.

## Adım 5: API İstemcisi ve Tipler
- `api/client.ts`: `axios` instance, `baseURL` olarak `import.meta.env.VITE_API_BASE_URL` ve JWT için `Authorization` header (localStorage’daki token) ekleyen interceptor.
- `api/*`: Auth, Chatbot, Source, Analytics için placeholder metodlar (listeleme, oluşturma, güncelleme, silme iskeleti).
- `types/*`: Auth (User, Token), Chatbot (Chatbot, Source), API (PaginatedResponse vs.) temel tipler.
- `utils/localStorage.ts`: Token set/get/remove yardımcıları.

## Adım 6: Veri Yönetimi (Query) ve Hooklar
- Query: Güncel paket adı ile `@tanstack/react-query` kurulacak (dokümandaki `react-query` ifadesi yerine). `QueryClient` ve `QueryClientProvider` `main.tsx` içine.
- `hooks/useAuth.ts`: login/logout akışını, token saklamayı ve kullanıcı bilgisini expose eden iskelet.
- `hooks/useChatbots.ts`: chatbot listesi/getir/fetch invalidate gibi iskelet sorgular.
- `hooks/usePagination.ts`: sayfalama durum yönetimi için basit yardımcı.

## Adım 7: Doğrulama ve Çalıştırma
- `npm install` sonrası `npm run dev` ile yerel geliştirme sunucusu başlatılacak.
- Ana rota ve layout’ların render olduğunu, Tailwind’in çalıştığını, Query ve Axios kurulumlarının derlenebildiğini doğrulama.
- `.env.development` dosyasına `VITE_API_BASE_URL` değeri eklenerek axios instance’ın doğru çalıştığı kontrol edilecek (backend mevcutsa basit health check çağrısı yapılabilir).

## Notlar ve Varsayımlar
- `react-query` paketinin güncel adı `@tanstack/react-query`; güncel ekosistemle uyum için bu paket kullanılacak.
- shadcn UI, paket değil CLI bileşen üreticisidir; ihtiyaç olan bileşenler gerektikçe eklenecek. Başlangıçta `shared` altına Tailwind tabanlı basit bileşenler yazılacak, sonradan shadcn ile zenginleştirilecek.
- Radix bileşenleri özelleştirilebilir; yalnızca kullanılan bileşenlerin paketleri kurulacak.
- TS geçişinde Vite JS template’inden başlıyoruz; iskelet dosyalar TS/TSX olarak oluşturulacak ve `tsconfig.json` ayarlanacak.

Onayınızla birlikte bu planı adım adım uygulayıp projeyi ayağa kaldıracağım.