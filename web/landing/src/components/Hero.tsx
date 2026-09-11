import { useEffect, useRef } from 'react'
import QRCode from 'qrcode'
import { usePrefersReducedMotion } from '../hooks/usePrefersReducedMotion'

const DEMO_URL = 'http://sharely.local/bluebird/'

export function Hero() {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const reducedMotion = usePrefersReducedMotion()

  useEffect(() => {
    if (!canvasRef.current) return
    QRCode.toCanvas(canvasRef.current, DEMO_URL, {
      width: 108,
      margin: 0,
      color: {
        dark: '#0a0a0c',
        light: '#ffffff',
      },
    }).catch(() => {
      /* canvas unavailable, fail silently */
    })
  }, [])

  const delay = (ms: number) => (reducedMotion ? undefined : { animationDelay: `${ms}ms` })

  return (
    <section id="top" className="mx-auto max-w-5xl px-5 pb-16 pt-14 sm:px-6 sm:pt-20">
      <div className="grid gap-12 lg:grid-cols-[1.05fr_1fr] lg:items-center lg:gap-10">
        <div>
          <h1 className="text-[2.35rem] font-semibold leading-[1.12] tracking-tight sm:text-[2.75rem]">
            Put this folder on the network. Right now.
          </h1>
          <p className="mt-5 max-w-md text-[1.05rem] leading-relaxed text-[var(--text-muted)]">
            <code className="rounded bg-[var(--code-bg)] px-1.5 py-0.5 font-mono text-[0.95em]">
              cd
            </code>{' '}
            into a project,{' '}
            <code className="rounded bg-[var(--code-bg)] px-1.5 py-0.5 font-mono text-[0.95em]">
              sharely start
            </code>
            , and it&rsquo;s reachable from any phone or laptop on your Wi-Fi. No build step, no
            deploy, no account.
          </p>
          <div className="mt-8 flex flex-wrap items-center gap-3">
            <a
              href="#install"
              className="rounded-md bg-[var(--accent)] px-4 py-2.5 text-sm font-medium text-[var(--accent-text-on)] transition-colors hover:bg-[var(--accent-hover)]"
            >
              Get Sharely
            </a>
            <a
              href="#"
              className="rounded-md border border-[var(--border-strong)] px-4 py-2.5 text-sm font-medium transition-colors hover:border-[var(--text-faint)]"
            >
              View on GitHub
            </a>
          </div>

          <div className="mt-10 flex flex-wrap items-center gap-x-2 gap-y-2 text-[13px] text-[var(--text-faint)]">
            <span className="rounded border border-[var(--border)] px-2 py-1 font-mono">
              cd my-project
            </span>
            <Arrow />
            <span className="rounded border border-[var(--border)] px-2 py-1 font-mono">
              sharely start
            </span>
            <Arrow />
            <span className="rounded border border-[var(--border)] px-2 py-1 font-mono">
              sharely.local/bluebird
            </span>
            <Arrow />
            <span aria-hidden="true">📱</span>
            <span className="sr-only">a phone on the same network</span>
          </div>
        </div>

        <div className="relative">
          <div className="overflow-hidden rounded-lg border border-[var(--border-strong)] bg-[var(--terminal-bg)] shadow-sm">
            <div className="flex items-center gap-1.5 border-b border-white/10 px-4 py-2.5">
              <span className="h-2.5 w-2.5 rounded-full bg-[#5f5f5f]" />
              <span className="h-2.5 w-2.5 rounded-full bg-[#5f5f5f]" />
              <span className="h-2.5 w-2.5 rounded-full bg-[#5f5f5f]" />
              <span className="ml-2 font-mono text-[11px] text-[var(--terminal-muted)]">
                ~/projects/bluebird
              </span>
            </div>
            <div className="px-5 py-5 font-mono text-[13px] leading-[1.65] text-[var(--terminal-text)] sm:text-[13.5px]">
              <p>
                <span className="text-[var(--terminal-accent)]">$</span> sharely start
              </p>
              <p
                className={reducedMotion ? '' : 'animate-[fadeIn_0.4s_ease-out_both]'}
                style={delay(150)}
              >
                &nbsp;
              </p>
              <p
                className={reducedMotion ? '' : 'animate-[fadeIn_0.4s_ease-out_both]'}
                style={delay(150)}
              >
                Sharely
              </p>
              <p
                className={reducedMotion ? '' : 'animate-[fadeIn_0.4s_ease-out_both]'}
                style={delay(300)}
              >
                &nbsp;
              </p>
              <p
                className={reducedMotion ? '' : 'animate-[fadeIn_0.4s_ease-out_both]'}
                style={delay(300)}
              >
                Sharing&nbsp;&nbsp;./bluebird
              </p>
              <p
                className={reducedMotion ? '' : 'animate-[fadeIn_0.4s_ease-out_both]'}
                style={delay(500)}
              >
                &nbsp;
              </p>
              <p
                className={(reducedMotion ? '' : 'animate-[fadeIn_0.4s_ease-out_both] ') + 'text-[var(--terminal-accent)]'}
                style={delay(500)}
              >
                ✓ Ready.
              </p>
              <p
                className={reducedMotion ? '' : 'animate-[fadeIn_0.4s_ease-out_both]'}
                style={delay(700)}
              >
                &nbsp;
              </p>
              <p
                className={reducedMotion ? '' : 'animate-[fadeIn_0.4s_ease-out_both]'}
                style={delay(700)}
              >
                &nbsp;&nbsp;sharely.local/bluebird
              </p>
              <p
                className={reducedMotion ? '' : 'animate-[fadeIn_0.4s_ease-out_both]'}
                style={delay(850)}
              >
                &nbsp;
              </p>
              <p
                className={(reducedMotion ? '' : 'animate-[fadeIn_0.4s_ease-out_both] ') + 'text-[var(--terminal-muted)]'}
                style={delay(850)}
              >
                &nbsp;&nbsp;Available for 1 hour
              </p>

              <div
                className={
                  'mt-4 flex items-center gap-4 ' +
                  (reducedMotion ? '' : 'animate-[fadeIn_0.5s_ease-out_both]')
                }
                style={delay(1050)}
              >
                <div className="rounded bg-white p-1.5">
                  <canvas ref={canvasRef} width={108} height={108} role="img" aria-label={`QR code linking to ${DEMO_URL}`} />
                </div>
                <p className="text-[var(--terminal-muted)]">
                  &nbsp;&nbsp;Scan to open
                </p>
              </div>
            </div>
          </div>
          <p className="mt-3 text-center text-[12px] text-[var(--text-faint)] sm:hidden">
            Real terminal output, real QR code &mdash; scan it with your phone.
          </p>
        </div>
      </div>

      <style>{`
        @keyframes fadeIn {
          from { opacity: 0; transform: translateY(2px); }
          to { opacity: 1; transform: translateY(0); }
        }
      `}</style>
    </section>
  )
}

function Arrow() {
  return (
    <span aria-hidden="true" className="text-[var(--text-faint)]">
      &rarr;
    </span>
  )
}
