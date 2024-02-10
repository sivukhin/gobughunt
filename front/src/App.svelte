<script lang="ts">
    import {Link, Route, Router} from "svelte-routing";
    import Dashboard from "./lib/Dashboard.svelte";
    import About from "./lib/About.svelte";
    import Tasks from "./lib/Tasks.svelte";
    import Highlights from "./lib/Highlights.svelte";

    export let url = "";
</script>

<Router url={url}>
    <main>
        <header>
            <h1>gobughunter</h1>
            <nav>
                |
                <Link to="/">dashboard</Link>
                |
                <Link to="/tasks">lint tasks</Link>
                |
                <Link to="/about">about</Link>
                |
            </nav>
        </header>
        <Route path="/" component={Dashboard}/>
        <Route path="/tasks" component={Tasks}/>
        <Route path="/about" component={About}/>
        <Route path="/highlights/:lintId" let:params>
            <Highlights lintId="{params.lintId}"/>
        </Route>
    </main>
</Router>

<style>
    main {
        width: 100%;
    }

    header {
        display: flex;
        flex-direction: row;
        justify-content: space-between;
        place-items: baseline;
        flex-wrap: wrap;
    }
</style>
