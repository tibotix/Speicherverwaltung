package main

import (
	"fmt"
	"os"
	"math/rand"
)

var mmu_debug = false

type TLBEntry struct {
	page_start uint16
	page_frame_start uint16
	valid bool
}

type TLB struct {
	entries [32]TLBEntry
	head int
}

// page table entry
type PTE struct {
	present bool
	dirty bool
	accessed bool
	page_frame_start uint16
	address_in_swap uint16
}

// page directory entry
type PDE struct {
	present bool
	dirty bool
	accessed bool
	page_table [8]PTE
}

type MMU struct {
	PAGE_SIZE uint16
	cr3 [8]PDE // cpu control register 3
	tlb TLB
	memory       []float64
	page_index_head uint16
	swap           []float64
}

func log_page_fault(vm_page_start uint16) {
	f, err := os.OpenFile("./page_fault_log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
	    panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("page fault requesting VM aligned page(%d)\n", vm_page_start)); err != nil {
	    panic(err)
	}
}

func create_mmu_instance() MMU {
	var page_directory [8]PDE
	for i := 0; i<8; i++ {
		var page_table [8]PTE
		for x := 0; x<8; x++ {
			var address_in_swap uint16 = uint16(((i*8)+x)*1024)
			pte := PTE{present: false, dirty: false, accessed: false, address_in_swap: address_in_swap}
			page_table[x] = pte
		}
		page_directory[i] = PDE{present: false, dirty: false, accessed: false, page_table: page_table}
	}
	tlb := TLB{head: 0}
	mmu := MMU{PAGE_SIZE: 1024, cr3: page_directory, tlb: tlb, memory: make([]float64, 4096), swap: make([]float64, 65536)}
	mmu.cr3[0].page_table[0].present = true
	mmu.cr3[0].page_table[0].page_frame_start = 0

	mmu.cr3[0].page_table[1].present = true
	mmu.cr3[0].page_table[1].page_frame_start = 1024

	mmu.cr3[0].page_table[2].present = true
	mmu.cr3[0].page_table[2].page_frame_start = 2048

	mmu.cr3[0].page_table[3].present = true
	mmu.cr3[0].page_table[3].page_frame_start = 3072

	return mmu
}

func (tlb *TLB) update_tlb(page_start uint16, page_frame_start uint16) {
	// We are using a simple FIFO algorithm for TLB replacement policy
	if mmu_debug {
		fmt.Println("Update TLB - Page (",page_start,") <-> Page frame (",page_frame_start,") TLB_HEAD: ", tlb.head)
		fmt.Scanln()
	}
	// check if there is already an entry for page_start 
	// NOTE: we can do this better using hashmaps, but im too lazy right now:)
	for idx, entry := range tlb.entries {
		if entry.page_frame_start == page_frame_start {
			// Invalidate all old tlb entries
			tlb.entries[idx].valid = false // go copies value in range for loop (so we have to use the idx here)
		}
		if entry.page_start == page_start {
			tlb.entries[idx].page_frame_start = page_frame_start
			tlb.entries[idx].valid = true
			tlb.head += 1
			return
		}
	}
	// no entry exists, overwrite new one
	tlb.entries[tlb.head] = TLBEntry{page_start: page_start, page_frame_start: page_frame_start, valid: true}
	tlb.head = (tlb.head+1) % len(tlb.entries)
}

func (mmu *MMU) select_page_to_replace() *PTE {
	// We are using the Second-chance (FIFO with reference bit) algorithm for page replacement
	for {
		pd_index := (mmu.page_index_head>>3) & 7
		pt_index := mmu.page_index_head & 7
		pte := &mmu.cr3[pd_index].page_table[pt_index]

		if pte.present && !pte.accessed {
			return pte
		}

		pte.accessed = false
		mmu.page_index_head = (mmu.page_index_head+1)%64
	}
}

func (mmu *MMU) select_page_to_replace_random() *PTE {
	pte_index := rand.Intn(4) // 4 possible present pages
	for pd_index, pde := range mmu.cr3 {
		for pt_index, pte := range pde.page_table {
			if pte.present {
				if pte_index == 0{
					return &mmu.cr3[pd_index].page_table[pt_index]
				}
				pte_index -= 1
			}
		}
	}
	panic("FATAL: could not find a suitable page to replace")
}


func (mmu *MMU) copy_page_to_swap(pte PTE){
	if mmu_debug {
		fmt.Println("Copy page VM(",pte.page_frame_start,") to SWAP(",pte.address_in_swap,")")
	}
	var i uint16 = 0
	for i = 0; i < mmu.PAGE_SIZE; i++ {
		mmu.swap[pte.address_in_swap+i] = mmu.memory[pte.page_frame_start+i]
	}
}

func (mmu *MMU) copy_page_from_swap(pte_in_swap PTE, pte_in_pm PTE){
	if mmu_debug {
		fmt.Println("Copy page in SWAP(",pte_in_swap.address_in_swap,") to PM(",pte_in_pm.page_frame_start,")")
	}
	var i uint16 = 0
	for i = 0; i < mmu.PAGE_SIZE; i++ {
		mmu.memory[pte_in_pm.page_frame_start+i] = mmu.swap[pte_in_swap.address_in_swap+i]
	}
}

func (mmu *MMU) vm_to_pte(vm_address uint16) *PTE {
	// calculate pte starting from cr3 -> pde -> pte
	pd_index := (vm_address>>13) & 7
	pt_index := (vm_address>>10) & 7
	pde := &mmu.cr3[pd_index]
	pte := &pde.page_table[pt_index]
	return pte
}

func (mmu *MMU) vm_to_pm(vm_address uint16) uint16 {
	vm_page_start := ^(mmu.PAGE_SIZE-1) & vm_address
	vm_page_offset := (mmu.PAGE_SIZE-1) & vm_address

	// 1. Check tlb
	for _, tlb_entry := range mmu.tlb.entries {
		if tlb_entry.valid && tlb_entry.page_start == vm_page_start {
			// Matching entry found
			if mmu_debug{
				fmt.Println("tlb_entry for vm: " + fmt.Sprintf("%d", vm_page_start) + "found with page_frame: " + fmt.Sprintf("%d", tlb_entry.page_frame_start))
			}
			return tlb_entry.page_frame_start + vm_page_offset
		}
	}

	// TLB MISS!
	if mmu_debug{
		fmt.Println("TLB MISS!")
		mmu.tlb.print_tlb()
	}
	// we have to simulate cpu page table walk starting from cr3
	pte := mmu.vm_to_pte(vm_address)
	if pte.present {
		// page found and page is present. Update flags, update tlb and return
		pte.accessed = true
		mmu.tlb.update_tlb(vm_page_start, pte.page_frame_start)
		return pte.page_frame_start + vm_page_offset
	}

	// PAGE FAULT! bring page frame to pm and update tlb
	if mmu_debug{
		fmt.Println("PAGE FAULT! vm_page_start: " + fmt.Sprintf("%d", vm_page_start))
	}
	log_page_fault(vm_page_start)
	pte_to_swap := mmu.select_page_to_replace()
	//pte_to_swap := mmu.select_page_to_replace_random()
	if pte_to_swap.dirty {
		mmu.copy_page_to_swap(*pte_to_swap)
		pte.dirty = false
	}
	mmu.copy_page_from_swap(*pte, *pte_to_swap)
	pte_to_swap.present = false
	pte.present = true
	pte.page_frame_start = pte_to_swap.page_frame_start
	mmu.tlb.update_tlb(vm_page_start, pte.page_frame_start)

	return mmu.vm_to_pm(vm_address)
}


func (mmu *MMU) print_mmu() {
	for pd_index, pde := range mmu.cr3 {
		for pt_index, pte := range pde.page_table {
			fmt.Println("PD_INDEX: ",pd_index, "PT_INDEX: ", pt_index, "SWAP: ", pte.address_in_swap, "PF_START: ", pte.page_frame_start, "PRESENT: ", pte.present, "ACCESSED: ", pte.accessed)
		}
	}
	fmt.Scanln()
}

func (tlb *TLB) print_tlb() {
	for idx, entry := range tlb.entries {
		fmt.Println("TLB_INDEX: ", idx, " VM(",entry.page_start,")", " <-> (",entry.page_frame_start,")", " VALID: ",entry.valid)
	}
}

func (mmu *MMU) read(index int) float64 {
	value := mmu.memory[mmu.vm_to_pm(uint16(index))]
	if mmu_debug {
		fmt.Println("Load ("+fmt.Sprintf("%d", index)+") = "+fmt.Sprintf("%f", value))
	}
	return value
}

func (mmu *MMU) write(index int, value float64){
	if mmu_debug {
		fmt.Println("Write (" + fmt.Sprintf("%d", index) + ") = " + fmt.Sprintf("%f", value))
	}
	mmu.memory[mmu.vm_to_pm(uint16(index))] = value
	mmu.vm_to_pte(uint16(index)).dirty = true
}