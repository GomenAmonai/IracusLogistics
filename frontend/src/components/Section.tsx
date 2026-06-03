import type { ReactNode } from 'react'

import { useReveal } from '../lib/hooks'

type SectionProps = {
  id?: string
  eyebrow?: string
  title?: ReactNode
  intro?: ReactNode
  children: ReactNode
  surface?: boolean
}

// Секция лендинга с reveal-набеганием и опциональной бумажной подложкой.
export function Section({ id, eyebrow, title, intro, children, surface }: SectionProps) {
  const { ref, isVisible } = useReveal<HTMLElement>()

  return (
    <section
      id={id}
      ref={ref}
      className={`reveal ${isVisible ? 'reveal-in' : ''} ${
        surface ? 'bg-paper-raised' : ''
      } border-t border-rule`}
    >
      <div className="mx-auto w-full max-w-6xl px-5 py-20 sm:px-8 md:py-28">
        {(eyebrow || title || intro) && (
          <header className="mb-12 max-w-2xl md:mb-16">
            {eyebrow && <p className="eyebrow mb-4">{eyebrow}</p>}
            {title && (
              <h2 className="font-display text-3xl font-extrabold leading-[1.05] tracking-[-0.02em] text-ink sm:text-4xl md:text-5xl">
                {title}
              </h2>
            )}
            {intro && (
              <p className="mt-5 text-base leading-relaxed text-ink-soft sm:text-lg">
                {intro}
              </p>
            )}
          </header>
        )}
        {children}
      </div>
    </section>
  )
}
