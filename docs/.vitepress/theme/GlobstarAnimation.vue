<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'

// Move constants outside component
const TIMING = {
  COLLAPSE: 800,
  EXPAND: 400,
  CHAR_DELAY: 30,
  PRE_HIGHLIGHT: 600,
  HIGHLIGHT_DURATION: 400,
  POST_HIGHLIGHT: 800,
  PAUSE: 1000,
  MIN_CHAR_DISPLAY: 100 // minimum ms per character to ensure readability
} as const

const LOGO = '( * )' as const

// Add proper TypeScript interfaces
interface Expression {
  code: string
  highlight: string
}

interface CharState {
  char: string
  id: string
  highlighted: boolean
  visible: boolean
  opacity: number
  transform: string
  width: string
}

const props = defineProps<{
  expressions?: Expression[]
}>()

// Use default props more efficiently
const expressions = computed(() => props.expressions ?? [
  {
    code: '(if_statement condition: (tuple (integer) (integer)))',
    highlight: '(tuple (integer) (integer))'
  },
  {
    code: '(binary_expression left: (integer) @left operator: "is" right: (integer) @right)',
    highlight: 'operator: "is"'
  },
  {
    code: '(call_expression function: (member_expression object: "console" property: "log"))',
    highlight: 'property: "log"'
  },
  {
    code: '(try_statement (catch_clause parameter: (identifier) @err body: (return_statement)))',
    highlight: '(return_statement)'
  },
  {code: '(function_definition parameters: (parameters (default_parameter value: (list))))',
    highlight: '(list)'
  },
  {
    code: '(if_statement consequence: (return_statement) alternative: (else_clause))',
    highlight: '(else_clause)'
  }
])

const chars = ref<CharState[]>([])
const isAnimating = ref(false)
const currentExpressionIndex = ref(-1)
let isTransitioning = false

// Convert string to array of character objects
const stringToChars = (str: string, options: Partial<CharState> = {}): CharState[] => {
  return str.split('').map((char, index) => ({
    char,
    id: `${index}-${char}-${Date.now()}`,
    highlighted: false,
    visible: true,
    opacity: options.opacity ?? 1,
    transform: options.transform ?? 'scale(1)',
    width: options.width ?? '1ch'
  }))
}

const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

const findMatchingParens = (expr: string) => {
  const matches: [number, number][] = []
  const stack: number[] = []
  
  for (let i = 0; i < expr.length; i++) {
    if (expr[i] === '(') {
      stack.push(i)
    } else if (expr[i] === ')') {
      const openIndex = stack.pop()
      if (openIndex !== undefined) {
        matches.push([openIndex, i + 1])
      }
    }
  }
  
  return matches
}

const findAllOccurrences = (str: string, substr: string) => {
  const ranges: [number, number][] = []
  let pos = 0
  
  while (pos < str.length) {
    const index = str.indexOf(substr, pos)
    if (index === -1) break
    ranges.push([index, index + substr.length])
    pos = index + 1
  }
  
  return ranges
}

// Collapse expression to ( * )
const collapseToLogo = async () => {
  if (!chars.value.length) return

  const currentLength = chars.value.length
  const centerIndex = Math.floor(currentLength / 2)

  // Batch DOM updates
  const newChars = chars.value.map((char, i) => {
    const distanceFromCenter = Math.abs(i - centerIndex)
    const fadeOut = distanceFromCenter < currentLength / 4

    return {
      ...char,
      opacity: fadeOut ? 0 : char.opacity,
      width: fadeOut ? '0ch' : '1ch',
      transform: `translate3d(0,0,0) scale(${fadeOut ? 0.8 : 1})`
    }
  })

  chars.value = newChars
  await delay(TIMING.COLLAPSE / 3)

  chars.value = stringToChars(LOGO).map(char => ({
    ...char,
    opacity: 1,
    width: '1ch',
    transform: 'translate3d(0,0,0) scale(1)',
  }))

  await delay(TIMING.COLLAPSE / 3)
}

// Expand from ( * ) to expression
const expandFromLogo = async (expression: Expression) => {
  // First fade out the logo
  chars.value = chars.value.map(char => ({
    ...char,
    opacity: 0,
    transform: 'scale(0.9)'
  }))
  
  await delay(TIMING.EXPAND / 2)
  
  // Replace with expression chars
  chars.value = stringToChars(expression.code, {
    opacity: 0
  })
  
  // Fade in characters sequentially
  for (const element of chars.value) {
    element.opacity = 0.3
    await delay(TIMING.CHAR_DELAY)
  }
}

const getExpressionDisplayTime = (expression: Expression) => {
  // Calculate how long this expression needs to be displayed
  const charCount = expression.code.length
  return Math.max(
    TIMING.POST_HIGHLIGHT,
    charCount * TIMING.MIN_CHAR_DISPLAY
  )
}

const animateExpression = async (expression: Expression) => {
  if (isTransitioning) return
  isTransitioning = true
  isAnimating.value = true
  
  try {
    await delay(TIMING.PRE_HIGHLIGHT)
    
    // Reset all characters to dim state
    chars.value = chars.value.map(char => ({
      ...char,
      opacity: 0.3,
      highlighted: false
    }))
    
    // Handle parentheses highlighting
    const parenMatches = findMatchingParens(expression.code)
    for (const [start, end] of parenMatches) {
      for (let i = start; i < end; i++) {
        chars.value[i].opacity = 1
      }
    }
    
    const displayTime = getExpressionDisplayTime(expression)
    await delay(displayTime / 2)
    
    // Pattern highlight
    if (expression.highlight) {
      // Reset non-paren characters to dim
      chars.value = chars.value.map(char => ({
        ...char,
        opacity: 0.3
      }))
      
      const matches = findAllOccurrences(expression.code, expression.highlight)
      for (const [start, end] of matches) {
        for (let i = start; i < end; i++) {
          chars.value[i].opacity = 1
        }
      }
      
      await delay(displayTime / 2)
    }
    
  } finally {
    isTransitioning = false
    isAnimating.value = false
  }
}

const nextExpression = async () => {
  if (isTransitioning) return
  
  currentExpressionIndex.value = 
    (currentExpressionIndex.value + 1) % expressions.value.length
  const expression = expressions.value[currentExpressionIndex.value]
  
  await collapseToLogo()
  await delay(TIMING.PAUSE)
  await expandFromLogo(expression)
  await animateExpression(expression)
}

const startAnimation = () => {
  const runNextAnimation = async () => {
    const currentExpression = expressions.value[
      (currentExpressionIndex.value + 1) % expressions.value.length
    ]
    const displayTime = getExpressionDisplayTime(currentExpression)
    
    await nextExpression()
    await delay(TIMING.PAUSE)
    
    // Schedule next animation
    setTimeout(runNextAnimation, displayTime + TIMING.COLLAPSE + TIMING.EXPAND)
  }

  setTimeout(runNextAnimation, 1000)
}

// Initialize with logo
chars.value = stringToChars(LOGO)

onMounted(() => {
  setTimeout(startAnimation, 1000)
})

// Update cleanup to handle the new animation approach
onUnmounted(() => {
  const highestTimeoutId = window.setTimeout(() => {}, 0) // skipcq: JS-0321
  for (let i = 0; i < highestTimeoutId; i++) {
    clearTimeout(i)
  }
})
</script>

<template>
  <div 
    class="globstar-logo" 
    :class="{ 'is-animating': isAnimating }"
    aria-hidden="true"
  >
    <pre><code><span
      v-for="char in chars"
      :key="char.id"
      class="char"
      :style="{
        opacity: char.opacity,
        transform: char.transform,
        width: char.width
      }"
    >{{ char.char }}</span></code></pre>
  </div>
</template>

<style scoped>
.globstar-logo {
  position: relative;
  font-family: monospace;
  font-size: 1.5rem;
  padding: 2rem;
  display: inline-block;
  will-change: transform;
  opacity: 0.7;
}

@media (max-width: 768px) {
    .globstar-logo {
        font-size: 1.1rem;
    }
}

.globstar-logo pre {
  margin: 0;
  padding: 0;
}

.globstar-logo code {
  font-family: inherit;
  white-space: pre;
}

.char {
  display: inline-block;
  transition: all 1s cubic-bezier(0.25, 0.1, 0.25, 1.4);
  will-change: transform, opacity;
  backface-visibility: hidden;
}

/* Dark mode adjustments */
@media (prefers-color-scheme: dark) {
  .char {
    opacity: 0.4;
  }
}
</style>