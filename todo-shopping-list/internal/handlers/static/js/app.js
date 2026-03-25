function initTodoSortable() {
    var el = document.getElementById('todo-sortable');
    if (!el) return;
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

function initShoppingSortable() {
    var el = document.getElementById('shopping-sortable');
    if (!el) return;
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
                modal.className = 'fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4';
                modal.innerHTML = '<div class="bg-white rounded-xl shadow-xl w-full max-w-lg p-6" onclick="event.stopPropagation()" id="edit-shopping-content"></div>';
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
    var html = '<h2 class="text-lg font-semibold text-gray-900 mb-4">Edit Category</h2>' +
        '<form hx-put="settings/categories/' + id + '" hx-target="#categories-list" hx-swap="innerHTML"' +
        ' hx-on::after-request="if(event.detail.successful) document.getElementById(\'edit-modal\').classList.add(\'hidden\')">' +
        '<div class="space-y-4">' +
        '<div><label class="block text-sm font-medium text-gray-700 mb-1">Name</label>' +
        '<input type="text" name="name" value="' + name + '" required class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"></div>' +
        '<div><label class="block text-sm font-medium text-gray-700 mb-1">Color</label>' +
        '<input type="color" name="color" value="' + color + '" class="w-full h-10 rounded-lg border border-gray-300 cursor-pointer"></div>' +
        '<div><label class="block text-sm font-medium text-gray-700 mb-1">Icon</label>' +
        '<input type="text" name="icon" value="' + icon + '" maxlength="4" class="w-20 rounded-lg border border-gray-300 px-3 py-2 text-sm text-center focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"></div>' +
        '</div>' +
        '<div class="flex justify-end space-x-3 mt-6">' +
        '<button type="button" onclick="document.getElementById(\'edit-modal\').classList.add(\'hidden\')" class="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors">Cancel</button>' +
        '<button type="submit" class="px-4 py-2 text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 rounded-lg transition-colors">Save</button>' +
        '</div></form>';
    showEditModal(html);
}

function editTag(id, name, color) {
    var html = '<h2 class="text-lg font-semibold text-gray-900 mb-4">Edit Tag</h2>' +
        '<form hx-put="settings/tags/' + id + '" hx-target="#tags-list" hx-swap="innerHTML"' +
        ' hx-on::after-request="if(event.detail.successful) document.getElementById(\'edit-modal\').classList.add(\'hidden\')">' +
        '<div class="space-y-4">' +
        '<div><label class="block text-sm font-medium text-gray-700 mb-1">Name</label>' +
        '<input type="text" name="name" value="' + name + '" required class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"></div>' +
        '<div><label class="block text-sm font-medium text-gray-700 mb-1">Color</label>' +
        '<input type="color" name="color" value="' + color + '" class="w-full h-10 rounded-lg border border-gray-300 cursor-pointer"></div>' +
        '</div>' +
        '<div class="flex justify-end space-x-3 mt-6">' +
        '<button type="button" onclick="document.getElementById(\'edit-modal\').classList.add(\'hidden\')" class="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors">Cancel</button>' +
        '<button type="submit" class="px-4 py-2 text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 rounded-lg transition-colors">Save</button>' +
        '</div></form>';
    showEditModal(html);
}

function showEditModal(html) {
    var modal = document.getElementById('edit-modal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'edit-modal';
        modal.className = 'fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4';
        modal.innerHTML = '<div class="bg-white rounded-xl shadow-xl w-full max-w-md p-6" onclick="event.stopPropagation()" id="edit-modal-content"></div>';
        modal.addEventListener('click', function(e) {
            if (e.target === modal) modal.classList.add('hidden');
        });
        document.body.appendChild(modal);
    }
    document.getElementById('edit-modal-content').innerHTML = html;
    modal.classList.remove('hidden');
    htmx.process(document.getElementById('edit-modal-content'));
}

// Shopping history autocomplete via plain fetch (avoids hx-boost conflicts)
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
