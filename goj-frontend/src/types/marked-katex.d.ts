declare module 'marked-katex-extension' {
  interface KatexOptions {
    throwOnError?: boolean
    output?: string
    displayMode?: boolean
    strict?: boolean
    trust?: boolean
    macros?: Record<string, string>
    delimiters?: Array<{
      left: string
      right: string
      display: boolean
    }>
    katex?: any
  }

  function markedKatex(options?: KatexOptions): any

  export default markedKatex
}
