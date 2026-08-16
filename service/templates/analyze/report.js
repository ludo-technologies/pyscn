(function () {
  var tabs = document.querySelectorAll('.tab');
  var panels = document.querySelectorAll('.panel');

  function show(id) {
    var found = false;
    panels.forEach(function (p) {
      var active = p.id === id;
      p.classList.toggle('active', active);
      found = found || active;
    });
    if (!found) { return; }
    tabs.forEach(function (t) { t.classList.toggle('active', t.dataset.tab === id); });
    window.scrollTo({ top: 0 });
    try { history.replaceState(null, '', '#' + id); } catch (e) { /* sandboxed viewers may forbid this */ }
  }

  tabs.forEach(function (t) {
    t.addEventListener('click', function () { show(t.dataset.tab); });
  });
  document.querySelectorAll('[data-goto]').forEach(function (a) {
    a.addEventListener('click', function (e) { e.preventDefault(); show(a.dataset.goto); });
  });
  if (location.hash.length > 1) { show(location.hash.slice(1)); }
})();

function sortQualityTable(tableID, columnIndex, numeric, button) {
  var table = document.getElementById(tableID);
  if (!table) { return; }

  var body = table.tBodies[0];
  var rows = Array.prototype.slice.call(body.rows);
  var direction = button.dataset.direction === 'asc' ? 'desc' : 'asc';
  var multiplier = direction === 'asc' ? 1 : -1;

  rows.sort(function (left, right) {
    var leftValue = left.cells[columnIndex].dataset.sortValue || '';
    var rightValue = right.cells[columnIndex].dataset.sortValue || '';
    var comparison = numeric
      ? Number(leftValue) - Number(rightValue)
      : leftValue.localeCompare(rightValue);
    return comparison * multiplier;
  });
  rows.forEach(function (row) { body.appendChild(row); });

  table.querySelectorAll('th[aria-sort]').forEach(function (header) { header.removeAttribute('aria-sort'); });
  table.querySelectorAll('.table-sort').forEach(function (control) { delete control.dataset.direction; });
  button.dataset.direction = direction;
  button.parentElement.setAttribute('aria-sort', direction === 'asc' ? 'ascending' : 'descending');
}

function sortModuleQuality(columnIndex, numeric, button) {
  sortQualityTable('module-quality-table', columnIndex, numeric, button);
}

function sortDirectoryComplexity(columnIndex, numeric, button) {
  sortQualityTable('directory-complexity-table', columnIndex, numeric, button);
}
