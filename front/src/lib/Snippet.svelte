<script lang="ts">
    import {Highlight, LineNumbers} from "svelte-highlight";
    import go from "svelte-highlight/languages/go";
    import "svelte-highlight/styles/github.css";


    export let snippet: { code: string, startLine: number, endLine: number }
</script>

<Highlight language={go} code={snippet.code} let:highlighted>
    <LineNumbers
            {highlighted}
            hideBorder
            --padding-left="4px"
            --padding-right="4px"
            startingLineNumber={snippet.startLine}
    />
</Highlight>