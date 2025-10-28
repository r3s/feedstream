document.addEventListener("DOMContentLoaded", function () {
    const feedFilter = document.getElementById('feed-filter');
    const loadMoreBtn = document.getElementById("load-more-btn");
    const loading = document.getElementById("loading");
    const feedContent = document.getElementById("feed-content");

    if (feedFilter) {
        const savedFilter = localStorage.getItem('selectedFeed');
        if (savedFilter) {
            feedFilter.value = savedFilter;
            applyFeedFilter(savedFilter);
        }

        feedFilter.addEventListener('change', function() {
            const selectedFeed = this.value;
            localStorage.setItem('selectedFeed', selectedFeed);
            applyFeedFilter(selectedFeed);
        });
    }

    function applyFeedFilter(selectedFeed) {
        const feedItems = document.querySelectorAll('.feed-item');
        const dateSections = document.querySelectorAll('.date-section');
        
        feedItems.forEach(item => {
            const feedName = item.getAttribute('data-feed-name');
            if (selectedFeed === 'all' || feedName === selectedFeed) {
                item.style.display = '';
            } else {
                item.style.display = 'none';
            }
        });
        
        dateSections.forEach(section => {
            const visibleItems = section.querySelectorAll('.feed-item:not([style*="display: none"])');
            if (selectedFeed === 'all' || visibleItems.length > 0) {
                section.style.display = '';
            } else {
                section.style.display = 'none';
            }
        });
    }

    if (loadMoreBtn) {
        loadMoreBtn.addEventListener("click", function () {
            const nextOffset = this.getAttribute("data-next-offset");

            this.style.display = "none";
            loading.style.display = "block";

            fetch(`/feeds?days=${nextOffset}`)
                .then((response) => response.text())
                .then((html) => {
                    const parser = new DOMParser();
                    const doc = parser.parseFromString(html, "text/html");
                    const newContent = doc.querySelector("#feed-content");

                    if (newContent) {
                        const dateSections = newContent.querySelectorAll(".date-section");
                        const newLoadMoreSection = newContent.querySelector(".load-more-section");

                        dateSections.forEach((section) => {
                            feedContent.insertBefore(
                                section,
                                document.querySelector(".load-more-section")
                            );
                        });

                        const currentLoadMoreSection = document.querySelector(".load-more-section");
                        if (newLoadMoreSection) {
                            currentLoadMoreSection.replaceWith(newLoadMoreSection);
                            const newBtn = document.getElementById("load-more-btn");
                            if (newBtn) {
                                newBtn.addEventListener("click", arguments.callee);
                            }
                        } else {
                            currentLoadMoreSection.remove();
                        }
                    }
                })
                .catch((error) => {
                    console.error("Error loading more items:", error);
                    loading.style.display = "none";
                    loadMoreBtn.style.display = "block";
                    loadMoreBtn.textContent = "Error - Try Again";
                });
        });
    }
});