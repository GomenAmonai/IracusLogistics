import { useState } from 'react'

const NAV_LINKS = [
  { href: '#how', label: 'Как работаем' },
  { href: '#calc', label: 'Калькулятор' },
  { href: '#services', label: 'Услуги' },
  { href: '#faq', label: 'FAQ' },
  { href: '#contacts', label: 'Контакты' },
]

const PHONE = '+7 495 120-44-18'

export function Header() {
  const [isMenuOpen, setMenuOpen] = useState(false)

  return (
    <header className="sticky top-0 z-50 border-b border-line bg-base/85 backdrop-blur-md">
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between gap-4 px-5 sm:px-8">
        <a href="#top" className="flex items-center gap-2.5" aria-label="IcarisLogistics, на главную">
          <span
            aria-hidden="true"
            className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent font-display text-sm font-bold text-surface"
          >
            I
          </span>
          <span className="font-display text-lg font-extrabold tracking-[-0.01em] text-ink">
            Icaris
          </span>
          <span className="hidden font-mono text-xs tracking-wide text-ink-soft xs:inline">
            CN&nbsp;→&nbsp;RU
          </span>
        </a>

        <nav className="hidden items-center gap-7 md:flex" aria-label="Основная навигация">
          {NAV_LINKS.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="text-sm text-ink-soft transition-colors duration-200 hover:text-ink"
            >
              {link.label}
            </a>
          ))}
        </nav>

        <div className="flex items-center gap-3">
          <a
            href={`tel:${PHONE.replace(/[^+\d]/g, '')}`}
            className="hidden font-mono text-sm text-ink-soft transition-colors hover:text-ink lg:block"
          >
            {PHONE}
          </a>
          <a
            href="#lead"
            className="hidden rounded-full bg-accent px-5 py-2 text-sm font-semibold text-surface transition-colors duration-200 hover:bg-accent-deep sm:inline-block"
          >
            Рассчитать доставку
          </a>
          <button
            type="button"
            className="flex h-11 w-11 items-center justify-center rounded-lg border border-line text-ink transition-colors hover:border-accent md:hidden"
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
                Рассчитать доставку
              </a>
            </li>
          </ul>
        </nav>
      )}
    </header>
  )
}
