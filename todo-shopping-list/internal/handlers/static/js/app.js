function initTodoSortable() {
    var el = document.getElementById('todo-sortable');
    if (!el || Sortable.get(el)) return;
    new Sortable(el, {
        handle: '.drag-handle',
        animation: 150,
        ghostClass: 'sortable-ghost',
        chosenClass: 'sortable-chosen',
        onEnd: function () {
            var ids = Array.from(el.querySelectorAll('[data-id]')).map(function(item) {
                return item.getAttribute('data-id');
            });
            var listId = window.location.pathname.split('/').pop();
            fetch('todos/' + listId + '/reorder', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({item_ids: ids.map(Number)})
            });
        }
    });
}

// Drag and drop between the board's columns. Dropping saves the new order of
// every card on the board, then, if the card changed column, saves its status
// and re-renders the board from the server so everything reflects the result.
function initTodoBoard() {
    var board = document.getElementById('todo-board');
    if (!board) return;
    var listId = board.dataset.listId;
    board.querySelectorAll('.board-column').forEach(function (column) {
        if (Sortable.get(column)) return;
        new Sortable(column, {
            group: 'todo-board',
            draggable: '.board-card',
            filter: 'button',
            preventOnFilter: false,
            animation: 150,
            ghostClass: 'sortable-ghost',
            chosenClass: 'sortable-chosen',
            onEnd: function (evt) {
                var ids = Array.from(board.querySelectorAll('.board-card')).map(function (card) {
                    return Number(card.dataset.id);
                });
                var reorder = fetch('todos/' + encodeURIComponent(listId) + '/reorder', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({item_ids: ids})
                });
                if (evt.from === evt.to) return;
                var status = evt.to.dataset.status;
                reorder.catch(function () {}).then(function () {
                    htmx.ajax('POST', 'todos/items/' + encodeURIComponent(evt.item.dataset.id) + '/status', {
                        target: '#items-list',
                        swap: 'innerHTML',
                        values: {status: status}
                    });
                });
            }
        });
    });
}

function initShoppingSortable() {
    var el = document.getElementById('shopping-sortable');
    if (!el || Sortable.get(el)) return;
    new Sortable(el, {
        handle: '.drag-handle',
        animation: 150,
        ghostClass: 'sortable-ghost',
        chosenClass: 'sortable-chosen',
        onEnd: function () {
            var ids = Array.from(el.querySelectorAll('[data-id]')).map(function(item) {
                return item.getAttribute('data-id');
            });
            var listId = window.location.pathname.split('/').pop();
            fetch('shopping/' + listId + '/reorder', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({item_ids: ids.map(Number)})
            });
        }
    });
}

function selectHistoryItem(btn) {
    var nameInput = document.getElementById('item-name-input');
    var form = nameInput.closest('form');
    nameInput.value = btn.dataset.name;
    var qtyInput = form.querySelector('input[name="quantity"]');
    if (qtyInput && btn.dataset.quantity) {
        qtyInput.value = btn.dataset.quantity;
    }
    var unitInput = form.querySelector('input[name="unit"]');
    if (unitInput && btn.dataset.unit) {
        unitInput.value = btn.dataset.unit;
    }
    document.getElementById('history-results').classList.add('hidden');
}

function editTodoItem(itemId) {
    fetch('todos/items/' + itemId)
        .then(function(r) { return r.text(); })
        .then(function(html) {
            document.getElementById('edit-item-content').innerHTML = html;
            document.getElementById('edit-item-modal').classList.remove('hidden');
            htmx.process(document.getElementById('edit-item-content'));
        });
}

function editShoppingItem(itemId) {
    fetch('shopping/items/' + itemId)
        .then(function(r) { return r.text(); })
        .then(function(html) {
            var modal = document.getElementById('edit-shopping-modal');
            if (!modal) {
                modal = document.createElement('div');
                modal.id = 'edit-shopping-modal';
                modal.className = 'fixed inset-0 bg-[#11111B]/70 flex items-center justify-center z-50 p-4';
                modal.innerHTML = '<div class="bg-[#313244] rounded-xl shadow-xl w-full max-w-lg p-6 border border-[#45475A]" onclick="event.stopPropagation()" id="edit-shopping-content"></div>';
                modal.addEventListener('click', function(e) {
                    if (e.target === modal) modal.classList.add('hidden');
                });
                document.body.appendChild(modal);
            }
            document.getElementById('edit-shopping-content').innerHTML = html;
            modal.classList.remove('hidden');
            htmx.process(document.getElementById('edit-shopping-content'));
        });
}

function editCategory(id, name, color, icon) {
    var html = '<h2 class="text-lg font-semibold text-[#CDD6F4] mb-4">Edit Category</h2>' +
        '<form hx-put="settings/categories/' + encodeURIComponent(id) + '" hx-target="#categories-list" hx-swap="innerHTML"' +
        ' hx-on::after-request="if(event.detail.successful) document.getElementById(\'edit-modal\').classList.add(\'hidden\')">' +
        '<div class="space-y-4">' +
        '<div><label class="block text-sm font-medium text-[#BAC2DE] mb-1">Name</label>' +
        '<input type="text" name="name" required class="w-full rounded-lg border border-[#585B70] bg-[#1E1E2E] px-3 py-2 text-sm text-[#CDD6F4] focus:ring-2 focus:ring-[#89B4FA] focus:border-[#89B4FA] outline-none"></div>' +
        '<div><label class="block text-sm font-medium text-[#BAC2DE] mb-1">Color</label>' +
        '<input type="color" name="color" class="w-full h-10 rounded-lg border border-[#585B70] bg-[#1E1E2E] cursor-pointer"></div>' +
        '<div><label class="block text-sm font-medium text-[#BAC2DE] mb-1">Icon</label>' +
        '<input type="text" name="icon" maxlength="4" class="w-20 rounded-lg border border-[#585B70] bg-[#1E1E2E] px-3 py-2 text-sm text-center text-[#CDD6F4] focus:ring-2 focus:ring-[#89B4FA] focus:border-[#89B4FA] outline-none"></div>' +
        '</div>' +
        '<div class="flex justify-end space-x-3 mt-6">' +
        '<button type="button" onclick="document.getElementById(\'edit-modal\').classList.add(\'hidden\')" class="px-4 py-2 text-sm font-medium text-[#BAC2DE] hover:bg-[#45475A] rounded-lg transition-colors">Cancel</button>' +
        '<button type="submit" class="px-4 py-2 text-sm font-medium text-[#11111B] bg-[#89B4FA] hover:bg-[#74C7EC] rounded-lg transition-colors">Save</button>' +
        '</div></form>';
    showEditModal(html, {name: name, color: color, icon: icon});
}

function editTag(id, name, color) {
    var html = '<h2 class="text-lg font-semibold text-[#CDD6F4] mb-4">Edit Tag</h2>' +
        '<form hx-put="settings/tags/' + encodeURIComponent(id) + '" hx-target="#tags-list" hx-swap="innerHTML"' +
        ' hx-on::after-request="if(event.detail.successful) document.getElementById(\'edit-modal\').classList.add(\'hidden\')">' +
        '<div class="space-y-4">' +
        '<div><label class="block text-sm font-medium text-[#BAC2DE] mb-1">Name</label>' +
        '<input type="text" name="name" required class="w-full rounded-lg border border-[#585B70] bg-[#1E1E2E] px-3 py-2 text-sm text-[#CDD6F4] focus:ring-2 focus:ring-[#89B4FA] focus:border-[#89B4FA] outline-none"></div>' +
        '<div><label class="block text-sm font-medium text-[#BAC2DE] mb-1">Color</label>' +
        '<input type="color" name="color" class="w-full h-10 rounded-lg border border-[#585B70] bg-[#1E1E2E] cursor-pointer"></div>' +
        '</div>' +
        '<div class="flex justify-end space-x-3 mt-6">' +
        '<button type="button" onclick="document.getElementById(\'edit-modal\').classList.add(\'hidden\')" class="px-4 py-2 text-sm font-medium text-[#BAC2DE] hover:bg-[#45475A] rounded-lg transition-colors">Cancel</button>' +
        '<button type="submit" class="px-4 py-2 text-sm font-medium text-[#11111B] bg-[#89B4FA] hover:bg-[#74C7EC] rounded-lg transition-colors">Save</button>' +
        '</div></form>';
    showEditModal(html, {name: name, color: color});
}

// values are assigned through the value property rather than interpolated into
// the markup, so names containing quotes or tags cannot inject HTML.
function showEditModal(html, values) {
    var modal = document.getElementById('edit-modal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'edit-modal';
        modal.className = 'fixed inset-0 bg-[#11111B]/70 flex items-center justify-center z-50 p-4';
        modal.innerHTML = '<div class="bg-[#313244] rounded-xl shadow-xl w-full max-w-md p-6 border border-[#45475A]" onclick="event.stopPropagation()" id="edit-modal-content"></div>';
        modal.addEventListener('click', function(e) {
            if (e.target === modal) modal.classList.add('hidden');
        });
        document.body.appendChild(modal);
    }
    var content = document.getElementById('edit-modal-content');
    content.innerHTML = html;
    if (values) {
        Object.keys(values).forEach(function(field) {
            var input = content.querySelector('[name="' + field + '"]');
            if (input) input.value = values[field];
        });
    }
    modal.classList.remove('hidden');
    htmx.process(content);
}

(function() {
    var timer = null;
    document.addEventListener('input', function(e) {
        if (e.target.id !== 'item-name-input') return;
        var q = e.target.value.trim();
        var results = document.getElementById('history-results');
        if (!results) return;

        clearTimeout(timer);
        if (q.length === 0) {
            results.classList.add('hidden');
            results.innerHTML = '';
            return;
        }
        timer = setTimeout(function() {
            fetch('shopping/history/search?name=' + encodeURIComponent(q))
                .then(function(r) { return r.text(); })
                .then(function(html) {
                    results.innerHTML = html;
                    if (html.trim()) {
                        results.classList.remove('hidden');
                    } else {
                        results.classList.add('hidden');
                    }
                });
        }, 300);
    });
})();

// Bind drag and drop whenever content appears: htmx.onLoad runs on the first
// page load and again after every swap. Inline scripts in the partials could
// not do this, because on a full page load they ran before this file defined
// the init functions. Each init skips elements that are already bound, so the
// repeat runs are harmless, and the flag stops hx-boost navigations (which
// re-run this file) from registering the hook again.
if (!window.sortableHookInstalled) {
    window.sortableHookInstalled = true;
    htmx.onLoad(function () {
        if (typeof Sortable === 'undefined') return;
        initTodoSortable();
        initShoppingSortable();
        initTodoBoard();
    });
}
