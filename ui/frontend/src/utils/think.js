import { ThinkTags } from "../constants/chat";

/**
 * Parses the thinking content from a raw text string.
 * It handles <think> and </think> tags and returns an array of segments
 * in the order they appear in the text.
 * 
 * @param {string} text The raw text content to parse
 * @returns {Object} An object containing the segments and state
 */
export function parseThinkContent(text) {
  const raw = String(text || "");
  const openTag = ThinkTags.Open;
  const closeTag = ThinkTags.Close;
  
  const segments = [];
  let i = 0;
  let inThink = false;
  let currentThink = "";
  let currentAnswer = "";

  while (i < raw.length) {
    const openAt = raw.indexOf(openTag, i);
    const closeAt = raw.indexOf(closeTag, i);

    let nextTagAt = -1;
    let isOpening = false;

    if (openAt !== -1 && (closeAt === -1 || openAt < closeAt)) {
      nextTagAt = openAt;
      isOpening = true;
    } else if (closeAt !== -1) {
      nextTagAt = closeAt;
      isOpening = false;
    }

    if (nextTagAt === -1) {
      const remaining = raw.slice(i);
      
      // Check for partial tags at the end of stream
      let partialLen = 0;
      for (let len = openTag.length - 1; len > 0; len--) {
        if (remaining.endsWith(openTag.slice(0, len))) {
          partialLen = len;
          break;
        }
      }
      if (partialLen === 0) {
        for (let len = closeTag.length - 1; len > 0; len--) {
          if (remaining.endsWith(closeTag.slice(0, len))) {
            partialLen = len;
            break;
          }
        }
      }

      const stableContent = remaining.slice(0, remaining.length - partialLen);
      if (inThink) currentThink += stableContent;
      else currentAnswer += stableContent;
      
      break;
    }

    const beforeTag = raw.slice(i, nextTagAt);
    if (inThink) currentThink += beforeTag;
    else currentAnswer += beforeTag;

    // Push the collected content before switching state
    if (inThink && currentThink) {
      segments.push({ type: 'think', content: currentThink.trim() });
      currentThink = "";
    } else if (!inThink && currentAnswer) {
      segments.push({ type: 'text', content: currentAnswer });
      currentAnswer = "";
    }

    if (isOpening) {
      inThink = true;
      i = nextTagAt + openTag.length;
    } else {
      inThink = false;
      i = nextTagAt + closeTag.length;
    }
  }

  // Final segments
  if (inThink && currentThink) {
    segments.push({ type: 'think', content: currentThink.trim() });
  } else if (!inThink && currentAnswer) {
    segments.push({ type: 'text', content: currentAnswer });
  }

  return {
    segments,
    isThinkingOpen: inThink,
    // For backward compatibility if needed, though we should prefer segments
    think: segments.filter(s => s.type === 'think').map(s => s.content).join("\n\n"),
    answer: segments.filter(s => s.type === 'text').map(s => s.content).join(""),
    hasAnswer: segments.some(s => s.type === 'text' && s.content.trim().length > 0)
  };
}
