import { useEffect, useState } from 'react'

const NAV_LINKS = [
  { href: '#how', label: 'Как работаем' },
  { href: '#calc', label: 'Расчёт' },
  { href: '#services', label: 'Услуги' },
  { href: '#faq', label: 'FAQ' },
  { href: '#contacts', label: 'Контакты' },
]

// Реальный телефон появится после регистрации юрлица — до тех пор контакт = Telegram менеджера.
const MANAGER_TG = '@hikill8'
const MANAGER_TG_URL = 'https://t.me/hikill8'

export function Header() {
  const [isMenuOpen, setMenuOpen] = useState(false)
  // Поверх тёмного hero шапка прозрачная со светлым текстом; на скролле — светлый солид.
  const [isScrolled, setScrolled] = useState(false)

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 24)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  const solid = isScrolled || isMenuOpen

  return (
    <header
      className={`fixed top-0 z-50 w-full border-b transition-colors duration-300 ${
        solid ? 'border-line bg-base/85 backdrop-blur-md' : 'border-transparent bg-transparent'
      }`}
    >
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between gap-4 px-5 sm:px-8">
        <a href="#top" className="flex items-center gap-2.5" aria-label="IcarisLogistics, на главную">
          <span
            aria-hidden="true"
            className={`flex h-8 w-8 items-center justify-center rounded-lg font-display text-sm font-bold transition-colors ${
              solid ? 'bg-accent text-surface' : 'bg-amber text-night'
            }`}
          >
            I
          </span>
          <span
            className={`font-display text-lg font-extrabold tracking-[-0.01em] transition-colors ${
              solid ? 'text-ink' : 'text-white'
            }`}
          >
            Icaris
          </span>
          <span
            className={`hidden font-mono text-xs tracking-wide transition-colors xs:inline ${
              solid ? 'text-ink-soft' : 'text-white/60'
            }`}
          >
            CN&nbsp;→&nbsp;RU
          </span>
        </a>

        <nav className="hidden items-center gap-7 md:flex" aria-label="Основная навигация">
          {NAV_LINKS.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className={`text-sm transition-colors duration-200 ${
                solid ? 'text-ink-soft hover:text-ink' : 'text-white/75 hover:text-white'
              }`}
            >
              {link.label}
            </a>
          ))}
        </nav>

        <div className="flex items-center gap-3">
          <a
            href={MANAGER_TG_URL}
            target="_blank"
            rel="noreferrer"
            className={`hidden font-mono text-sm transition-colors lg:block ${
              solid ? 'text-ink-soft hover:text-ink' : 'text-white/75 hover:text-white'
            }`}
          >
            {MANAGER_TG}
          </a>
          <a
            href="#lead"
            className={`hidden rounded-full px-5 py-2 text-sm font-semibold transition-colors duration-200 sm:inline-block ${
              solid
                ? 'bg-accent text-surface hover:bg-accent-deep'
                : 'bg-amber text-night hover:bg-white'
            }`}
          >
            Запросить расчёт
          </a>
          <button
            type="button"
            className={`flex h-11 w-11 items-center justify-center rounded-lg border transition-colors md:hidden ${
              solid
                ? 'border-line text-ink hover:border-accent'
                : 'border-white/25 text-white hover:border-white/50'
            }`}
            aria-expanded={isMenuOpen}
            aria-controls="mobile-nav"
            aria-label={isMenuOpen ? 'Закрыть меню' : 'Открыть меню'}
            onClick={() => setMenuOpen((open) => !open)}
          >
            <span aria-hidden="true" className="text-lg leading-none">
              {isMenuOpen ? '✕' : '≡'}
            </span>
          </button>
        </div>
      </div>

      {isMenuOpen && (
        <nav
          id="mobile-nav"
          className="border-t border-line bg-base px-5 py-4 md:hidden"
          aria-label="Мобильная навигация"
        >
          <ul className="flex flex-col gap-1">
            {NAV_LINKS.map((link) => (
              <li key={link.href}>
                <a
                  href={link.href}
                  className="block rounded-lg px-2 py-3 text-base text-ink-soft transition-colors hover:bg-surface hover:text-ink"
                  onClick={() => setMenuOpen(false)}
                >
                  {link.label}
                </a>
              </li>
            ))}
            <li>
              <a
                href="#lead"
                className="mt-2 block rounded-full bg-accent px-2 py-3 text-center text-base font-semibold text-surface"
                onClick={() => setMenuOpen(false)}
              >
                Запросить расчёт
              </a>
            </li>
          </ul>
        </nav>
      )}
    </header>
  )
}
