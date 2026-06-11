// TODO(hiki): реквизиты — плейсхолдеры; заменить реальными ИНН/ОГРН/адресом до публикации.
const REQUISITES = [
  { label: 'Юр. лицо', value: 'ООО «Иракус Логистикс»' },
  { label: 'ИНН', value: '7701234567' },
  { label: 'ОГРН', value: '1187700001234' },
  { label: 'Адрес', value: 'Москва, ул. Складочная, 6с1, офис 412' },
]

// Телефон/почта появятся с реальными после регистрации юрлица — фейковые не публикуем.
const CONTACTS = [
  { label: 'Менеджер', value: '@hikill8', href: 'https://t.me/hikill8' },
  { label: 'Бот', value: '@IcarisLogBot', href: 'https://t.me/IcarisLogBot' },
]

// Числовые реквизиты идут на терминал-чип; адрес/название — обычный текст.
const TERMINAL_REQUISITES = new Set(['ИНН', 'ОГРН'])

export function Footer() {
  return (
    <footer id="contacts" className="border-t border-line bg-base">
      <div className="mx-auto w-full max-w-6xl px-5 py-16 sm:px-8">
        <div className="grid gap-10 md:grid-cols-[1fr_1fr_auto]">
          <div>
            {/* Бренд-марка как в шапке: кобальтовый скруглённый квадрат-литера + Icaris */}
            <div className="flex items-center gap-2.5">
              <span
                aria-hidden="true"
                className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent font-display text-sm font-bold text-surface"
              >
                I
              </span>
              <span className="font-display text-lg font-extrabold tracking-[-0.01em] text-ink">
                Icaris
              </span>
            </div>
            <p className="mt-4 max-w-xs text-sm leading-relaxed text-ink-soft">
              Экспедирование грузов из Китая в Россию: выкуп, консолидация, таможня и
              доставка до двери под договором.
            </p>
          </div>

          {/* Реквизиты — чистые строки реестра с тихими делителями; числа на терминал-чипе */}
          <dl className="text-sm">
            {REQUISITES.map((item) => (
              <div
                key={item.label}
                className="flex flex-col gap-1 border-b border-line-soft py-3 first:border-t xs:flex-row xs:items-center xs:gap-3"
              >
                <dt className="shrink-0 text-[0.7rem] uppercase tracking-[0.06em] text-ink-soft xs:w-24">
                  {item.label}
                </dt>
                <dd>
                  {TERMINAL_REQUISITES.has(item.label) ? (
                    <span className="terminal text-sm">{item.value}</span>
                  ) : (
                    <span className="tabular font-mono text-ink">{item.value}</span>
                  )}
                </dd>
              </div>
            ))}
          </dl>

          <div className="flex flex-col gap-5">
            <ul className="space-y-3 text-sm">
              {CONTACTS.map((contact) => (
                <li key={contact.label} className="flex flex-col">
                  <span className="text-[0.7rem] uppercase tracking-[0.06em] text-ink-soft">
                    {contact.label}
                  </span>
                  <a
                    href={contact.href}
                    className="inline-flex min-h-11 items-center font-mono text-ink underline-offset-4 transition-colors duration-200 hover:text-accent hover:underline"
                  >
                    {contact.value}
                  </a>
                </li>
              ))}
            </ul>
            <a
              href="#lead"
              className="inline-flex items-center justify-center rounded-full bg-accent px-6 py-3 text-sm font-semibold text-surface shadow-card transition-colors duration-200 hover:bg-accent-deep"
            >
              Рассчитать доставку
            </a>
          </div>
        </div>

        <div className="mt-12 flex flex-col gap-2 border-t border-line pt-6 text-xs text-ink-soft sm:flex-row sm:items-center sm:justify-between">
          <span>© {new Date().getFullYear()} ООО «Иракус Логистикс». Все права защищены.</span>
          <span>Грузы под контролем</span>
        </div>
      </div>
    </footer>
  )
}
