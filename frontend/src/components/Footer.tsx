// TODO(hiki): реквизиты — плейсхолдеры; заменить реальными ИНН/ОГРН/адресом до публикации.
const REQUISITES = [
  { label: 'Юр. лицо', value: 'ООО «Иракус Логистикс»' },
  { label: 'ИНН', value: '7701234567' },
  { label: 'ОГРН', value: '1187700001234' },
  { label: 'Адрес', value: 'Москва, ул. Складочная, 6с1, офис 412' },
]

const CONTACTS = [
  { label: 'Телефон', value: '+7 495 120-44-18', href: 'tel:+74951204418' },
  { label: 'Почта', value: 'cargo@iracus.ru', href: 'mailto:cargo@iracus.ru' },
  { label: 'Telegram', value: '@iracus_logistics_bot', href: 'https://t.me/iracus_logistics_bot' },
]

export function Footer() {
  return (
    <footer id="contacts" className="border-t border-rule bg-paper">
      <div className="mx-auto w-full max-w-6xl px-5 py-16 sm:px-8">
        <div className="grid gap-10 md:grid-cols-[1fr_1fr_auto]">
          <div>
            {/* Бренд-марка как в шапке: квадрат-литера + Iracus */}
            <div className="flex items-center gap-2.5">
              <span
                aria-hidden="true"
                className="flex h-7 w-7 items-center justify-center border border-ink font-mono text-sm font-bold text-ink"
              >
                I
              </span>
              <span className="font-display text-lg font-extrabold uppercase tracking-[-0.01em] text-ink">
                Iracus
              </span>
            </div>
            <p className="mt-4 max-w-xs text-sm leading-relaxed text-ink-soft">
              Экспедирование грузов из Китая в Россию: выкуп, консолидация, таможня и
              доставка до двери под договором.
            </p>
          </div>

          {/* Реквизиты как строки реестра: ruled-разделители, mono-данные */}
          <dl className="border-t border-rule text-sm">
            {REQUISITES.map((item) => (
              <div key={item.label} className="flex gap-3 border-b border-rule py-2.5">
                <dt className="w-24 shrink-0 font-mono text-[0.7rem] uppercase tracking-[0.06em] text-ink-soft">
                  {item.label}
                </dt>
                <dd className="tabular font-mono text-ink">{item.value}</dd>
              </div>
            ))}
          </dl>

          <div className="flex flex-col gap-4">
            <ul className="space-y-2.5 text-sm">
              {CONTACTS.map((contact) => (
                <li key={contact.label} className="flex flex-col">
                  <span className="font-mono text-[0.7rem] uppercase tracking-[0.06em] text-ink-soft">
                    {contact.label}
                  </span>
                  <a
                    href={contact.href}
                    className="inline-flex min-h-11 items-center font-mono text-ink underline-offset-4 transition-colors duration-200 hover:underline"
                  >
                    {contact.value}
                  </a>
                </li>
              ))}
            </ul>
            <a
              href="#lead"
              className="inline-flex items-center justify-center bg-stamp px-5 py-3 font-mono text-sm font-semibold uppercase tracking-[0.06em] text-paper transition-colors duration-200 hover:bg-stamp-deep"
            >
              Рассчитать доставку
            </a>
          </div>
        </div>

        <div className="mt-12 flex flex-col gap-2 border-t border-rule pt-6 font-mono text-xs uppercase tracking-[0.04em] text-ink-soft sm:flex-row sm:items-center sm:justify-between">
          <span>© {new Date().getFullYear()} ООО «Иракус Логистикс». Все права защищены.</span>
          <span>Грузы под контролем</span>
        </div>
      </div>
    </footer>
  )
}
