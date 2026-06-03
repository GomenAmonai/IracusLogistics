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
    <header className="sticky top-0 z-50 border-b border-rule bg-paper/90 backdrop-blur-md">
      <div className="mx-auto flex h-[4.5rem] w-full max-w-6xl items-center justify-between gap-4 px-5 sm:px-8">
        <a href="#top" className="flex items-center gap-2.5" aria-label="Iracus Logistics, на главную">
          <span
            aria-hidden="true"
            className="flex h-7 w-7 items-center justify-center border border-ink font-mono text-sm font-bold text-ink"
          >
            I
          </span>
          <span className="font-display text-lg font-extrabold uppercase tracking-[-0.01em] text-ink">
            Iracus
          </span>
          <span className="hidden font-mono text-xs tracking-[0.1em] text-ink-soft xs:inline">
            CN→RU
          </span>
        </a>

        <nav className="hidden items-center gap-7 md:flex" aria-label="Основная навигация">
          {NAV_LINKS.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="font-mono text-xs uppercase tracking-[0.06em] text-ink-soft transition-colors duration-200 hover:text-ink"
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
            className="hidden bg-stamp px-4 py-2 font-mono text-xs font-semibold uppercase tracking-[0.06em] text-paper transition-colors duration-200 hover:bg-stamp-deep sm:inline-block"
          >
            Рассчитать
          </a>
          <button
            type="button"
            className="flex h-11 w-11 items-center justify-center border border-rule text-ink transition-colors hover:border-ink md:hidden"
            aria-expanded={isMenuOpen}
            aria-controls="mobile-nav"
            aria-label={isMenuOpen ? 'Закрыть меню' : 'Открыть меню'}
            onClick={() => setMenuOpen((open) => !open)}
          >
            <span aria-hidden="true" className="font-mono text-lg leading-none">
              {isMenuOpen ? '✕' : '≡'}
            </span>
          </button>
        </div>
      </div>

      {isMenuOpen && (
        <nav
          id="mobile-nav"
          className="border-t border-rule bg-paper px-5 py-4 md:hidden"
          aria-label="Мобильная навигация"
        >
          <ul className="flex flex-col gap-1">
            {NAV_LINKS.map((link) => (
              <li key={link.href}>
                <a
                  href={link.href}
                  className="block px-2 py-3 font-mono text-sm uppercase tracking-[0.06em] text-ink-soft transition-colors hover:bg-paper-sunk hover:text-ink"
                  onClick={() => setMenuOpen(false)}
                >
                  {link.label}
                </a>
              </li>
            ))}
            <li>
              <a
                href="#lead"
                className="mt-2 block bg-stamp px-2 py-3 text-center font-mono text-sm font-semibold uppercase tracking-[0.06em] text-paper"
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
