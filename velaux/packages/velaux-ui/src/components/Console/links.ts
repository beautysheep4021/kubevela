// splitLinks parts a row's links into those shown and those folded under
// More: the first few show, and a lone link left over shows in place of a
// More holding only it. Hidden links are dropped first, and a link marked
// more always folds, as one too long for the column.
export function splitLinks<T extends { hidden?: boolean; more?: boolean }>(
  links: T[],
  visible: number
): { shown: T[]; folded: T[] } {
  const kept = links.filter((l) => !l.hidden);
  const inline = kept.filter((l) => !l.more);
  const always = kept.filter((l) => l.more);
  if (always.length === 0 && inline.length <= visible + 1) {
    return { shown: inline, folded: [] };
  }
  if (always.length > 0 && inline.length <= visible) {
    return { shown: inline, folded: always };
  }
  return { shown: inline.slice(0, visible), folded: [...inline.slice(visible), ...always] };
}
