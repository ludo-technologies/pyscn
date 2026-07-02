package service

// communityGraphStyle holds the scoped CSS for the macro-architecture graph.
const communityGraphStyle = `<style>
.community-graph-help { margin: 4px 0 12px; color: #666; font-size: 13px; }
.community-graph { margin-bottom: 24px; }
.community-graph-controls {
    display: flex; flex-wrap: wrap; gap: 16px; align-items: center;
    margin-bottom: 8px; font-size: 13px; color: #333;
}
.community-graph-controls label { display: inline-flex; align-items: center; gap: 6px; }
.community-graph-controls select { padding: 2px 6px; }
.community-graph-count { color: #888; }
.community-graph-canvas-wrap { position: relative; }
#community-graph-canvas {
    width: 100%; max-width: 100%; height: 480px;
    border: 1px solid #e0e0e0; border-radius: 6px; background: #fafafa;
    display: block; cursor: grab; touch-action: none;
}
#community-graph-canvas.dragging { cursor: grabbing; }
.community-graph-tooltip {
    position: absolute; pointer-events: none; z-index: 10;
    background: rgba(33, 33, 33, 0.95); color: #fff; padding: 8px 10px;
    border-radius: 6px; font-size: 12px; line-height: 1.5; max-width: 280px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.3);
}
.community-graph-tooltip strong { color: #fff; }
.community-graph-tooltip .cg-bridge { color: #ffb3b3; font-weight: 600; }
.community-graph-legend {
    display: flex; flex-wrap: wrap; gap: 14px; margin-top: 10px; font-size: 12px; color: #444;
}
.community-graph-legend .cg-swatch {
    display: inline-block; width: 12px; height: 12px; border-radius: 3px;
    margin-right: 6px; vertical-align: -1px; border: 1px solid rgba(0,0,0,0.15);
}
.community-graph-legend .cg-item { display: inline-flex; align-items: center; }
</style>`

// communityGraphScript holds the self-contained, dependency-free renderer for
// the macro-architecture graph. It reads the embedded JSON blob, runs a small
// force-directed layout on a canvas, and wires up hover/click/filter behavior.
const communityGraphScript = `<script>
(function () {
    var dataEl = document.getElementById('community-graph-data');
    var canvas = document.getElementById('community-graph-canvas');
    if (!dataEl || !canvas) { return; }

    var data;
    try { data = JSON.parse(dataEl.textContent); } catch (e) { return; }
    if (!data || !data.nodes || !data.nodes.length) { return; }

    var ctx = canvas.getContext('2d');
    var tooltip = document.getElementById('community-graph-tooltip');
    var filterEl = document.getElementById('community-graph-filter');
    var bridgesEl = document.getElementById('community-graph-bridges');
    var isolatedEl = document.getElementById('community-graph-isolated');
    var countEl = document.getElementById('community-graph-count');
    var legendEl = document.getElementById('community-graph-legend');

    var BRIDGE_OUTLINE = '#d9534f';
    var CROSS_EDGE = 'rgba(217, 83, 79, 0.7)';
    var EDGE = 'rgba(120, 120, 120, 0.35)';

    var nodes = data.nodes.map(function (n, i) { return Object.assign({}, n, { idx: i, x: 0, y: 0, vx: 0, vy: 0 }); });
    var index = {};
    nodes.forEach(function (n) { index[n.id] = n; });
    var edges = data.edges.filter(function (e) { return index[e.from] && index[e.to]; });

    // Degree for radius scaling and isolated detection.
    nodes.forEach(function (n) { n.degree = 0; });
    edges.forEach(function (e) { index[e.from].degree++; index[e.to].degree++; });

    // Deterministic initial layout on a circle (avoids Math.random for stable renders).
    var N = nodes.length;
    nodes.forEach(function (n, i) {
        var a = (2 * Math.PI * i) / N;
        n.x = Math.cos(a) * 180;
        n.y = Math.sin(a) * 180;
    });

    function radius(n) {
        if (n.group) { return Math.max(10, Math.min(34, 8 + Math.sqrt(n.count || 1) * 3)); }
        return n.bridge ? 8 : 6;
    }

    // Populate the community filter dropdown.
    (data.communities || []).forEach(function (c) {
        var opt = document.createElement('option');
        opt.value = c.id;
        opt.textContent = c.id + ' (' + c.size + ')';
        filterEl && filterEl.appendChild(opt);
    });

    // Legend: one swatch per community plus bridge / cross-edge markers.
    if (legendEl) {
        var legendHTML = (data.communities || []).map(function (c) {
            var extra = c.dominantPackage ? (' · ' + c.dominantPackage) : (c.dominantLayer ? (' · ' + c.dominantLayer) : '');
            return '<span class="cg-item"><span class="cg-swatch" style="background:' + c.color + '"></span>' +
                escapeHtml(c.id + extra) + '</span>';
        }).join('');
        legendHTML += '<span class="cg-item"><span class="cg-swatch" style="background:#fff;border-color:' +
            BRIDGE_OUTLINE + ';border-width:2px"></span>Bridge module</span>';
        legendHTML += '<span class="cg-item"><span class="cg-swatch" style="background:' + CROSS_EDGE + '"></span>Cross-community edge</span>';
        legendEl.innerHTML = legendHTML;
    }

    function escapeHtml(s) {
        return String(s).replace(/[&<>"']/g, function (c) {
            return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
        });
    }

    // ---- visibility / filtering ----
    var visible = {};
    function recomputeVisible() {
        var community = filterEl ? filterEl.value : '';
        var bridgesOnly = bridgesEl ? bridgesEl.checked : false;
        var hideIsolated = isolatedEl ? isolatedEl.checked : false;

        nodes.forEach(function (n) {
            var show = true;
            if (community && n.community !== community) { show = false; }
            if (bridgesOnly && !n.bridge && !n.group) { show = false; }
            n._show = show;
        });
        if (hideIsolated) {
            var connected = {};
            edges.forEach(function (e) {
                if (index[e.from]._show && index[e.to]._show) { connected[e.from] = true; connected[e.to] = true; }
            });
            nodes.forEach(function (n) { if (!connected[n.id]) { n._show = false; } });
        }
        visible = {};
        var shown = 0;
        nodes.forEach(function (n) { if (n._show) { visible[n.id] = true; shown++; } });
        if (countEl) {
            countEl.textContent = shown + ' / ' + nodes.length + (data.collapsed ? ' communities' : ' modules');
        }
    }

    // ---- force simulation ----
    var alpha = 1;
    var selected = null;

    function tick() {
        var k = 6000; // repulsion strength
        for (var i = 0; i < N; i++) {
            var a = nodes[i];
            if (!a._show) { continue; }
            for (var j = i + 1; j < N; j++) {
                var b = nodes[j];
                if (!b._show) { continue; }
                var dx = a.x - b.x, dy = a.y - b.y;
                var d2 = dx * dx + dy * dy || 0.01;
                var f = k / d2;
                var d = Math.sqrt(d2);
                var fx = (dx / d) * f, fy = (dy / d) * f;
                a.vx += fx; a.vy += fy; b.vx -= fx; b.vy -= fy;
            }
        }
        edges.forEach(function (e) {
            var a = index[e.from], b = index[e.to];
            if (!a._show || !b._show) { return; }
            var dx = b.x - a.x, dy = b.y - a.y;
            var d = Math.sqrt(dx * dx + dy * dy) || 0.01;
            var target = 70;
            var f = (d - target) * 0.02;
            var fx = (dx / d) * f, fy = (dy / d) * f;
            a.vx += fx; a.vy += fy; b.vx -= fx; b.vy -= fy;
        });
        nodes.forEach(function (n) {
            if (!n._show) { return; }
            n.vx -= n.x * 0.002; // gravity to center
            n.vy -= n.y * 0.002;
            if (n === dragNode) { return; }
            n.x += n.vx * alpha; n.y += n.vy * alpha;
            n.vx *= 0.85; n.vy *= 0.85;
        });
        alpha *= 0.985;
        if (alpha < 0.02) { alpha = 0.02; }
    }

    // ---- rendering ----
    var dpr = window.devicePixelRatio || 1;
    var W = 800, H = 480, cx = 400, cy = 240, scale = 1;

    function resize() {
        var rect = canvas.getBoundingClientRect();
        W = rect.width || 800; H = rect.height || 480;
        canvas.width = W * dpr; canvas.height = H * dpr;
        ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
        cx = W / 2; cy = H / 2;
    }

    function neighborsOf(id) {
        var set = {};
        set[id] = true;
        edges.forEach(function (e) {
            if (e.from === id) { set[e.to] = true; }
            if (e.to === id) { set[e.from] = true; }
        });
        return set;
    }

    function draw() {
        ctx.clearRect(0, 0, W, H);
        var hl = selected ? neighborsOf(selected) : null;

        edges.forEach(function (e) {
            var a = index[e.from], b = index[e.to];
            if (!a._show || !b._show) { return; }
            var dim = hl && !(hl[e.from] && hl[e.to]);
            ctx.beginPath();
            ctx.moveTo(cx + a.x * scale, cy + a.y * scale);
            ctx.lineTo(cx + b.x * scale, cy + b.y * scale);
            ctx.strokeStyle = e.cross ? CROSS_EDGE : EDGE;
            ctx.globalAlpha = dim ? 0.08 : 1;
            ctx.lineWidth = e.weight ? Math.min(4, 1 + e.weight * 0.4) : (e.cross ? 1.5 : 1);
            ctx.stroke();
        });
        ctx.globalAlpha = 1;

        var showLabels = nodes.length <= 60 || data.collapsed;
        nodes.forEach(function (n) {
            if (!n._show) { return; }
            var dim = hl && !hl[n.id];
            var r = radius(n);
            var px = cx + n.x * scale, py = cy + n.y * scale;
            ctx.globalAlpha = dim ? 0.2 : 1;
            ctx.beginPath();
            ctx.arc(px, py, r, 0, 2 * Math.PI);
            ctx.fillStyle = n.bridge ? '#ffe2e2' : n.color;
            ctx.fill();
            ctx.lineWidth = n.bridge ? 2.5 : 1;
            ctx.strokeStyle = n.bridge ? BRIDGE_OUTLINE : 'rgba(0,0,0,0.25)';
            ctx.stroke();
            if (showLabels) {
                ctx.globalAlpha = dim ? 0.25 : 1;
                ctx.fillStyle = '#333';
                ctx.font = '11px -apple-system, sans-serif';
                ctx.textAlign = 'center';
                ctx.textBaseline = 'top';
                ctx.fillText(n.label, px, py + r + 2);
            }
        });
        ctx.globalAlpha = 1;
    }

    var running = true;
    function frame() {
        if (running) { tick(); }
        draw();
        requestAnimationFrame(frame);
    }

    function reheat() { alpha = Math.max(alpha, 0.6); running = true; }

    // ---- interaction ----
    var dragNode = null;
    function toGraph(evt) {
        var rect = canvas.getBoundingClientRect();
        return { x: (evt.clientX - rect.left - cx) / scale, y: (evt.clientY - rect.top - cy) / scale };
    }
    function nodeAt(evt) {
        var p = toGraph(evt);
        var best = null, bestD = Infinity;
        nodes.forEach(function (n) {
            if (!n._show) { return; }
            var dx = n.x - p.x, dy = n.y - p.y;
            var d = Math.sqrt(dx * dx + dy * dy);
            if (d < radius(n) + 4 && d < bestD) { best = n; bestD = d; }
        });
        return best;
    }

    canvas.addEventListener('mousemove', function (evt) {
        if (dragNode) {
            var p = toGraph(evt);
            dragNode.x = p.x; dragNode.y = p.y; dragNode.vx = 0; dragNode.vy = 0;
            reheat();
            return;
        }
        var n = nodeAt(evt);
        if (n && tooltip) {
            var rect = canvas.getBoundingClientRect();
            var html = '<strong>' + escapeHtml(n.label) + '</strong><br>Community: ' + escapeHtml(n.community);
            if (n.group) { html += '<br>Modules: ' + n.count; }
            if (n.bridge) {
                html += '<br><span class="cg-bridge">Bridge module</span>';
                if (n.crossEdges) { html += '<br>Cross edges: ' + n.crossEdges; }
                if (n.targets && n.targets.length) { html += '<br>Targets: ' + escapeHtml(n.targets.join(', ')); }
            }
            if (n.dominantPackage) { html += '<br>Package: ' + escapeHtml(n.dominantPackage); }
            if (n.dominantLayer) { html += '<br>Layer: ' + escapeHtml(n.dominantLayer); }
            tooltip.innerHTML = html;
            tooltip.hidden = false;
            var tx = evt.clientX - rect.left + 12, ty = evt.clientY - rect.top + 12;
            tooltip.style.left = Math.min(tx, W - 20) + 'px';
            tooltip.style.top = ty + 'px';
            canvas.style.cursor = 'pointer';
        } else if (tooltip) {
            tooltip.hidden = true;
            canvas.style.cursor = 'grab';
        }
    });
    canvas.addEventListener('mouseleave', function () { if (tooltip) { tooltip.hidden = true; } });
    canvas.addEventListener('mousedown', function (evt) {
        var n = nodeAt(evt);
        if (n) { dragNode = n; canvas.classList.add('dragging'); }
    });
    window.addEventListener('mouseup', function () {
        if (dragNode) { dragNode = null; canvas.classList.remove('dragging'); }
    });
    canvas.addEventListener('click', function (evt) {
        if (dragNode) { return; }
        var n = nodeAt(evt);
        selected = (n && selected !== n.id) ? n.id : null;
    });

    function applyFilters() { recomputeVisible(); reheat(); }
    filterEl && filterEl.addEventListener('change', applyFilters);
    bridgesEl && bridgesEl.addEventListener('change', applyFilters);
    isolatedEl && isolatedEl.addEventListener('change', applyFilters);

    window.addEventListener('resize', resize);

    resize();
    recomputeVisible();
    // The community graph may live inside a hidden tab; size it when shown.
    var wrap = document.getElementById('community-graph');
    if (wrap && 'IntersectionObserver' in window) {
        var io = new IntersectionObserver(function (entries) {
            entries.forEach(function (e) { if (e.isIntersecting) { resize(); reheat(); } });
        });
        io.observe(wrap);
    }
    requestAnimationFrame(frame);
})();
</script>`
