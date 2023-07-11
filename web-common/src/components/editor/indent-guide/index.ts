import {
  Decoration,
  DecorationSet,
  EditorView,
  ViewPlugin,
  WidgetType,
} from "@codemirror/view";
import IndentGuide from "./IndentGuide.svelte";

class IndentGuideWidget extends WidgetType {
  toDOM() {
    const element = document.createElement("span");
    element.style.display = "inline-block";
    element.style.position = "absolute";
    new IndentGuide({ target: element });
    return element;
  }
}

export const indentGuide = () =>
  ViewPlugin.fromClass(
    class {
      indentGuides: Decoration[];
      view: EditorView;
      decorations: DecorationSet;

      constructor(view) {
        this.view = view;
        this.indentGuides = this.calculateIndentGuides();
        this.decorations = this.createDecorations();
      }

      update(tr) {
        if (tr.docChanged) {
          this.indentGuides = this.calculateIndentGuides();
          this.decorations = this.createDecorations();
        }
      }

      /** Creates a Monaco-like indent */
      decorationsFromLine(lineNumber) {
        const line = this.view.state.doc.line(lineNumber);
        const indent = /^\s*/.exec(line.text)[0];
        const indentSize = indent.length;
        const decorations = [];

        for (let i = 0; i < indentSize; i++) {
          if (
            // tab
            indent[i] === "\t" ||
            // two spaces
            (indent[i] === " " && indent[i + 1] === " ") ||
            // case where we are indented, but user adds one additional space
            i === indentSize - 1
          ) {
            const guidePos = line.from + i;
            decorations.push(
              Decoration.widget({
                widget: new IndentGuideWidget(),
                side: -1,
              }).range(guidePos)
            );

            // If we have two spaces, skip the next space character
            if (indent[i] === " " && indent[i + 1] === " ") {
              i++;
            }
          }
        }

        return decorations;
      }

      calculateIndentGuides() {
        const guides = [];
        const lineCount = this.view.state.doc.lines;

        for (let i = 1; i <= lineCount; i++) {
          guides.push(...this.decorationsFromLine(i));
        }

        return guides;
      }

      createDecorations() {
        return Decoration.set(this.indentGuides);
      }
    },
    {
      decorations: (v) => v.decorations,
    }
  );
